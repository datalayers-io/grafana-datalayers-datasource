package arrow_flightsql

import (
	"strings"
	"testing"
	"time"

	"github.com/apache/arrow/go/v12/arrow"
	"github.com/apache/arrow/go/v12/arrow/array"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestNewQueryDataResponse(t *testing.T) {
	alloc := memory.DefaultAllocator
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "i8", Type: arrow.PrimitiveTypes.Int8},
			{Name: "i16", Type: arrow.PrimitiveTypes.Int16},
			{Name: "i32", Type: arrow.PrimitiveTypes.Int32},
			{Name: "i64", Type: arrow.PrimitiveTypes.Int64},
			{Name: "u8", Type: arrow.PrimitiveTypes.Uint8},
			{Name: "u16", Type: arrow.PrimitiveTypes.Uint16},
			{Name: "u32", Type: arrow.PrimitiveTypes.Uint32},
			{Name: "u64", Type: arrow.PrimitiveTypes.Uint64},
			{Name: "f32", Type: arrow.PrimitiveTypes.Float32},
			{Name: "f64", Type: arrow.PrimitiveTypes.Float64},
			{Name: "utf8", Type: &arrow.StringType{}},
			{Name: "duration", Type: &arrow.DurationType{}},
			{Name: "timestamp", Type: &arrow.TimestampType{}},
		},
		nil,
	)

	strValues := []jsonArray{
		newJSONArray(`[1, -2, 3]`, arrow.PrimitiveTypes.Int8),
		newJSONArray(`[1, -2, 3]`, arrow.PrimitiveTypes.Int16),
		newJSONArray(`[1, -2, 3]`, arrow.PrimitiveTypes.Int32),
		newJSONArray(`[1, -2, 3]`, arrow.PrimitiveTypes.Int64),
		newJSONArray(`[1, 2, 3]`, arrow.PrimitiveTypes.Uint8),
		newJSONArray(`[1, 2, 3]`, arrow.PrimitiveTypes.Uint16),
		newJSONArray(`[1, 2, 3]`, arrow.PrimitiveTypes.Uint32),
		newJSONArray(`[1, 2, 3]`, arrow.PrimitiveTypes.Uint64),
		newJSONArray(`[1.1, -2.2, 3.0]`, arrow.PrimitiveTypes.Float32),
		newJSONArray(`[1.1, -2.2, 3.0]`, arrow.PrimitiveTypes.Float64),
		newJSONArray(`["foo", "bar", "baz"]`, &arrow.StringType{}),
		newJSONArray(`[0, 1, -2]`, &arrow.DurationType{}),
		newJSONArray(`[0, 1, 2]`, &arrow.TimestampType{}),
	}

	var arr []arrow.Array
	for _, v := range strValues {
		tarr, _, err := array.FromJSON(
			alloc,
			v.dt,
			strings.NewReader(v.json),
		)
		require.NoError(t, err)
		arr = append(arr, tarr)
	}

	record := array.NewRecord(schema, arr, -1)
	records := []arrow.Record{record}
	reader, err := array.NewRecordReader(schema, records)
	require.NoError(t, err)

	query := sqlutil.Query{Format: sqlutil.FormatOptionTable}
	resp := newQueryDataResponse(errReader{RecordReader: reader}, query, metadata.MD{})
	require.NoError(t, resp.Error)
	require.Len(t, resp.Frames, 1)
	require.Len(t, resp.Frames[0].Fields, 13)

	expectedValues := [][]any{
		{any(int8(1)), any(int8(-2)), any(int8(3))},
		{any(int16(1)), any(int16(-2)), any(int16(3))},
		{any(int32(1)), any(int32(-2)), any(int32(3))},
		{any(int64(1)), any(int64(-2)), any(int64(3))},
		{any(uint8(1)), any(uint8(2)), any(uint8(3))},
		{any(uint16(1)), any(uint16(2)), any(uint16(3))},
		{any(uint32(1)), any(uint32(2)), any(uint32(3))},
		{any(uint64(1)), any(uint64(2)), any(uint64(3))},
		{any(float32(1.1)), any(float32(-2.2)), any(float32(3.0))},
		{any(float64(1.1)), any(float64(-2.2)), any(float64(3.0))},
		{any("foo"), any("bar"), any("baz")},
		{any(int64(0)), any(int64(1)), any(int64(-2))},
		{any(time.Unix(0, 0).UTC()), any(time.Unix(0, 1).UTC()), any(time.Unix(0, 2).UTC())},
	}

	frame := resp.Frames[0]
	for i, f := range frame.Fields {
		assert.Equal(t, schema.Field(i).Name, f.Name)
		switch schema.Field(i).Type.(type) {
		case *arrow.Int8Type:
			assert.Equal(t, data.FieldTypeInt8, f.Type())
		case *arrow.Int16Type:
			assert.Equal(t, data.FieldTypeInt16, f.Type())
		case *arrow.Int32Type:
			assert.Equal(t, data.FieldTypeInt32, f.Type())
		case *arrow.Int64Type:
			assert.Equal(t, data.FieldTypeInt64, f.Type())
		case *arrow.Uint8Type:
			assert.Equal(t, data.FieldTypeUint8, f.Type())
		case *arrow.Uint16Type:
			assert.Equal(t, data.FieldTypeUint16, f.Type())
		case *arrow.Uint32Type:
			assert.Equal(t, data.FieldTypeUint32, f.Type())
		case *arrow.Uint64Type:
			assert.Equal(t, data.FieldTypeUint64, f.Type())
		case *arrow.Float32Type:
			assert.Equal(t, data.FieldTypeFloat32, f.Type())
		case *arrow.Float64Type:
			assert.Equal(t, data.FieldTypeFloat64, f.Type())
		case *arrow.StringType:
			assert.Equal(t, data.FieldTypeString, f.Type())
		case *arrow.DurationType:
			assert.Equal(t, data.FieldTypeInt64, f.Type())
		case *arrow.TimestampType:
			assert.Equal(t, data.FieldTypeTime, f.Type())
		}
		assert.Equal(t, expectedValues[i], extractFieldValues(t, f))
	}
}

type jsonArray struct {
	json string
	dt   arrow.DataType
}

func newJSONArray(json string, dt arrow.DataType) jsonArray {
	return jsonArray{json: json, dt: dt}
}

type errReader struct {
	array.RecordReader
	err error
}

func (e errReader) Read() error {
	return e.err
}

func extractFieldValues(t *testing.T, field *data.Field) []any {
	t.Helper()
	values := make([]any, field.Len())
	for i := 0; i < field.Len(); i++ {
		values[i] = field.At(i)
	}
	return values
}
