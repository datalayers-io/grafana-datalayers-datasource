package arrow_flightsql

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
	"google.golang.org/grpc/metadata"
)

// queryRequest represents an inbound query request as part of a batch of queries sent to DataSource.QueryData.
type queryRequest struct {
	RefID                string `json:"refId"`
	Text                 string `json:"queryText"`
	IntervalMilliseconds int    `json:"intervalMs"`
	MaxDataPoints        int64  `json:"maxDataPoints"`
	Format               string `json:"format"`
}

// executeResult encapsulates concurrent query responses.
type executeResult struct {
	refID        string
	dataResponse backend.DataResponse
}

// QueryData executes a batch of ad-hoc queries and returns a batch of results.
func (d *DataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	executeResults := make(chan executeResult, len(req.Queries))
	var wg sync.WaitGroup

	for _, dataQuery := range req.Queries {
		query, err := decodeQueryRequest(dataQuery)
		if err != nil {
			response.Responses[dataQuery.RefID] = backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
			continue
		}
		// Check query.RawSQL, An empty query returns an empty array
    if query.RawSQL == "" {
      response.Responses[dataQuery.RefID] = backend.DataResponse{
        Frames: []*data.Frame{},
      }
      continue
    }

		wg.Add(1)
		go d.executeQuery(ctx, query, executeResults, &wg)
	}

	wg.Wait()
	close(executeResults)
	for result := range executeResults {
		response.Responses[result.refID] = result.dataResponse
	}

	return response, nil
}

// decodeQueryRequest decodes a backend.DataQuery and returns a sqlutil.Query with all macros expanded.
func decodeQueryRequest(dataQuery backend.DataQuery) (*sqlutil.Query, error) {
	var q queryRequest
	if err := json.Unmarshal(dataQuery.JSON, &q); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	format := formatQueryOptionFromString(q.Format)
	query := &sqlutil.Query{
		RawSQL:        q.Text,
		RefID:         q.RefID,
		MaxDataPoints: q.MaxDataPoints,
		Interval:      time.Duration(q.IntervalMilliseconds) * time.Millisecond,
		TimeRange:     dataQuery.TimeRange,
		Format:        format,
	}

	sql, err := sqlutil.Interpolate(query, macros)
	if err != nil {
		return nil, fmt.Errorf("interpolate macros: %w", err)
	}
	query.RawSQL = sql

	return query, nil
}

// executeQuery executes a single query in a goroutine and sends the result to the executeResults channel.
func (d *DataSource) executeQuery(ctx context.Context, query *sqlutil.Query, executeResults chan<- executeResult, wg *sync.WaitGroup) {
	defer wg.Done()
	executeResults <- executeResult{
		refID:        query.RefID,
		dataResponse: d.query(ctx, *query),
	}
}

// query executes a SQL statement by issuing a CommandStatementQuery command to Flight SQL.
func (d *DataSource) query(ctx context.Context, query sqlutil.Query) (response backend.DataResponse) {
	defer func(response *backend.DataResponse) {
		if r := recover(); r != nil {
			logErrorf("Panic: %s %s", r, string(debug.Stack()))
			*response = backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("panic: %s", r))
		}
	}(&response)

	if d.md.Len() != 0 {
		ctx = metadata.NewOutgoingContext(ctx, d.md)
	}

	info, err := d.client.Execute(ctx, query.RawSQL)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("flightsql: %s", err))
	}
	if len(info.Endpoint) != 1 {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("unsupported endpoint count in response: %d", len(info.Endpoint)))
	}
	reader, err := d.client.DoGetWithHeaderExtraction(ctx, info.Endpoint[0].Ticket)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("flightsql: %s", err))
	}
	defer reader.Release()

	headers, err := reader.Header()
	if err != nil {
		logErrorf("Failed to extract headers: %s", err)
	}

	return newQueryDataResponse(reader, query, headers)
}

// formatQueryOptionFromString returns the format query option based on the provided format string.
func formatQueryOptionFromString(format string) sqlutil.FormatQueryOption {
	switch format {
	case "table":
		return sqlutil.FormatOptionTable
	default:
		return sqlutil.FormatOptionTimeSeries
	}
}
