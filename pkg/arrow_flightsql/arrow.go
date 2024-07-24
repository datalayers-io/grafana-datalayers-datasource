package arrow_flightsql

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"time"

	"github.com/apache/arrow/go/v12/arrow"
	"github.com/apache/arrow/go/v12/arrow/array"
	"github.com/apache/arrow/go/v12/arrow/scalar"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
	"google.golang.org/grpc/metadata"
)

const rowLimit = 1_000_000

type recordReader interface {
	Next() bool
	Schema() *arrow.Schema
	Record() arrow.Record
	Err() error
}

// newQueryDataResponse builds a [backend.DataResponse] from a stream of
// [arrow.Record]s. The backend.DataResponse contains a single [data.Frame].
func newQueryDataResponse(reader recordReader, query sqlutil.Query, headers metadata.MD) backend.DataResponse {
	var resp backend.DataResponse
	frame, err := frameForRecords(reader)
	if err != nil {
		resp.Error = err
		return resp
	}

	if frame.Rows() == 0 {
		resp.Frames = data.Frames{}
		return resp
	}

	addFrameMetadata(frame, query, headers)
	formatFrameData(&resp, frame, query)

	return resp
}

func frameForRecords(reader recordReader) (*data.Frame, error) {
	frame := newFrame(reader.Schema())
	rows := int64(0)

	for reader.Next() {
		if err := appendRecordToFrame(frame, reader.Record()); err != nil {
			return nil, err
		}

		rows += reader.Record().NumRows()
		if rows > rowLimit {
			addRowLimitNotice(frame)
			return frame, nil
		}

		if err := reader.Err(); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return frame, nil
}

func newFrame(schema *arrow.Schema) *data.Frame {
	fields := schema.Fields()
	df := &data.Frame{
		Fields: make([]*data.Field, len(fields)),
		Meta:   &data.FrameMeta{},
	}
	for i, f := range fields {
		df.Fields[i] = newField(f)
	}
	return df
}

func newField(f arrow.Field) *data.Field {
	switch f.Type.ID() {
	case arrow.STRING:
		return newDataField[string](f)
	case arrow.FLOAT32:
		return newDataField[float32](f)
	case arrow.FLOAT64:
		return newDataField[float64](f)
	case arrow.UINT8:
		return newDataField[uint8](f)
	case arrow.UINT16:
		return newDataField[uint16](f)
	case arrow.UINT32:
		return newDataField[uint32](f)
	case arrow.UINT64:
		return newDataField[uint64](f)
	case arrow.INT8:
		return newDataField[int8](f)
	case arrow.INT16:
		return newDataField[int16](f)
	case arrow.INT32:
		return newDataField[int32](f)
	case arrow.INT64:
		return newDataField[int64](f)
	case arrow.BOOL:
		return newDataField[bool](f)
	case arrow.TIMESTAMP:
		return newDataField[time.Time](f)
	case arrow.DURATION:
		return newDataField[int64](f)
	default:
		return newDataField[json.RawMessage](f)
	}
}

func newDataField[T any](f arrow.Field) *data.Field {
	if f.Nullable {
		var s []*T
		return data.NewField(f.Name, nil, s)
	}
	var s []T
	return data.NewField(f.Name, nil, s)
}

func copyData(field *data.Field, col arrow.Array) error {
	defer recoverFromPanic()

	data := col.Data()
	switch col.DataType().ID() {
	case arrow.TIMESTAMP:
		return copyTimestampData(field, array.NewTimestampData(data))
	case arrow.DENSE_UNION:
		return copyDenseUnionData(field, array.NewDenseUnionData(data))
	default:
		return copyBasicData(field, data)
	}
}

func copyTimestampData(field *data.Field, data *array.Timestamp) error {
	for i := 0; i < data.Len(); i++ {
		if appendNullableTime(field, data, i) {
			continue
		}
		field.Append(data.Value(i).ToTime(arrow.Nanosecond))
	}
	return nil
}

func copyDenseUnionData(field *data.Field, data *array.DenseUnion) error {
	for i := 0; i < data.Len(); i++ {
		sc, err := scalar.GetScalar(data, i)
		if err != nil {
			return err
		}
		value := sc.(*scalar.DenseUnion).ChildValue()

		var jsonData any
		switch value.DataType().ID() {
		case arrow.STRING:
			jsonData = value.(*scalar.String).String()
		case arrow.BOOL:
			jsonData = value.(*scalar.Boolean).Value
		case arrow.INT32:
			jsonData = value.(*scalar.Int32).Value
		case arrow.INT64:
			jsonData = value.(*scalar.Int64).Value
		case arrow.LIST:
			jsonData = value.(*scalar.List).Value
		}
		b, err := json.Marshal(jsonData)
		if err != nil {
			return err
		}
		field.Append(json.RawMessage(b))
	}
	return nil
}

func copyBasicData(field *data.Field, data arrow.ArrayData) error {
	switch data.DataType().ID() {
	case arrow.STRING:
		copyBasic[string](field, array.NewStringData(data))
	case arrow.UINT8:
		copyBasic[uint8](field, array.NewUint8Data(data))
	case arrow.UINT16:
		copyBasic[uint16](field, array.NewUint16Data(data))
	case arrow.UINT32:
		copyBasic[uint32](field, array.NewUint32Data(data))
	case arrow.UINT64:
		copyBasic[uint64](field, array.NewUint64Data(data))
	case arrow.INT8:
		copyBasic[int8](field, array.NewInt8Data(data))
	case arrow.INT16:
		copyBasic[int16](field, array.NewInt16Data(data))
	case arrow.INT32:
		copyBasic[int32](field, array.NewInt32Data(data))
	case arrow.INT64:
		copyBasic[int64](field, array.NewInt64Data(data))
	case arrow.FLOAT32:
		copyBasic[float32](field, array.NewFloat32Data(data))
	case arrow.FLOAT64:
		copyBasic[float64](field, array.NewFloat64Data(data))
	case arrow.BOOL:
		copyBasic[bool](field, array.NewBooleanData(data))
	case arrow.DURATION:
		copyBasic[int64](field, array.NewInt64Data(data))
	}
	return nil
}

type arrowArray[T any] interface {
	IsNull(int) bool
	Value(int) T
	Len() int
}

func copyBasic[T any, Array arrowArray[T]](dst *data.Field, src Array) {
	for i := 0; i < src.Len(); i++ {
		if dst.Nullable() {
			if src.IsNull(i) {
				var s *T
				dst.Append(s)
				continue
			}
			s := src.Value(i)
			dst.Append(&s)
			continue
		}
		dst.Append(src.Value(i))
	}
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		logErrorf("Panic: %s %s", r, string(debug.Stack()))
	}
}

func appendRecordToFrame(frame *data.Frame, record arrow.Record) error {
	for i, col := range record.Columns() {
		if err := copyData(frame.Fields[i], col); err != nil {
			return err
		}
	}
	return nil
}

func addFrameMetadata(frame *data.Frame, query sqlutil.Query, headers metadata.MD) {
	frame.Meta.Custom = map[string]any{
		"headers": headers,
	}
	frame.Meta.ExecutedQueryString = query.RawSQL
	frame.Meta.DataTopic = data.DataTopic(query.RawSQL)
}

func formatFrameData(resp *backend.DataResponse, frame *data.Frame, query sqlutil.Query) {
	switch query.Format {
	case sqlutil.FormatOptionTimeSeries:
		formatTimeSeriesData(resp, frame)
	case sqlutil.FormatOptionTable:
		resp.Frames = data.Frames{frame}
	case sqlutil.FormatOptionLogs:
		resp.Frames = data.Frames{frame}
	default:
		resp.Error = fmt.Errorf("unsupported format")
	}
}

func formatTimeSeriesData(resp *backend.DataResponse, frame *data.Frame) {
	if _, idx := frame.FieldByName("time"); idx == -1 {
		resp.Error = fmt.Errorf("no time column found")
		return
	}

	if frame.TimeSeriesSchema().Type == data.TimeSeriesTypeLong {
		var err error
		frame, err = data.LongToWide(frame, nil)
		if err != nil {
			resp.Error = err
			return
		}
	}
	resp.Frames = data.Frames{frame}
}

func addRowLimitNotice(frame *data.Frame) {
	frame.AppendNotices(data.Notice{
		Severity: data.NoticeSeverityWarning,
		Text:     fmt.Sprintf("Results have been limited to %v because the SQL row limit was reached", rowLimit),
	})
}

func appendNullableTime(field *data.Field, data *array.Timestamp, i int) bool {
	if field.Nullable() {
		if data.IsNull(i) {
			var t *time.Time
			field.Append(t)
			return true
		}
		t := data.Value(i).ToTime(arrow.Nanosecond)
		field.Append(&t)
		return true
	}
	return false
}
