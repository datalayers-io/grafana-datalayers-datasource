package arrow_flightsql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/apache/arrow/go/v12/arrow/array"
	"github.com/apache/arrow/go/v12/arrow/flight"
	"github.com/apache/arrow/go/v12/arrow/flight/flightsql"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/go-chi/chi/v5"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
	"google.golang.org/grpc/metadata"
)

// Type assertions to ensure FlightSQLDatasource implements required interfaces
var (
	_ backend.QueryDataHandler      = (*DataSource)(nil)
	_ backend.CheckHealthHandler    = (*DataSource)(nil)
	_ instancemgmt.InstanceDisposer = (*DataSource)(nil)
	_ backend.CallResourceHandler   = (*DataSource)(nil)
)

// DataSource represents a Grafana datasource plugin for Flight SQL
type DataSource struct {
	client          *client
	resourceHandler backend.CallResourceHandler
	md              metadata.MD
}

// HTTP APIs

func (d *DataSource) getMacros(w http.ResponseWriter, r *http.Request) {
	size := len(sqlutil.DefaultMacros) + len(macros)
	names := make([]string, 0, size)
	for k := range sqlutil.DefaultMacros {
		if k == "table" || k == "column" {
			// We don't have the information available for these to function
			// propperly so omit them from advertisement.
			continue
		}
		names = append(names, k)
	}
	for k := range macros {
		names = append(names, k)
	}
	sort.Strings(names)

	err := json.NewEncoder(w).Encode(struct {
		Macros []string `json:"macros"`
	}{
		Macros: names,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (d *DataSource) getSQLInfo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, d.md)
	info, err := d.client.GetSqlInfo(ctx, []flightsql.SqlInfo{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reader, err := d.client.DoGet(ctx, info.Endpoint[0].Ticket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Release()

	if err := writeDataResponse(w, newDataResponse(reader)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (d *DataSource) getTables(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, d.md)

	info, err := d.client.GetTables(ctx, &flightsql.GetTablesOpts{
		TableTypes: []string{"BASE TABLE", "table"},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reader, err := d.client.DoGet(ctx, info.Endpoint[0].Ticket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Release()

	if err := writeDataResponse(w, newDataResponse(reader)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (d *DataSource) getColumns(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, `query parameter "table" is required`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, d.md)
	info, err := d.client.GetTables(ctx, &flightsql.GetTablesOpts{
		TableNameFilterPattern: &tableName,
		IncludeSchema:          true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reader, err := d.client.DoGet(ctx, info.Endpoint[0].Ticket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Release()

	if !reader.Next() {
		http.Error(w, "table not found", http.StatusNotFound)
		return
	}
	rec := reader.Record()
	rec.Retain()
	defer rec.Release()
	reader.Next()
	if err := reader.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indices := rec.Schema().FieldIndices("table_schema")
	if len(indices) == 0 {
		http.Error(w, "table_schema field not found", http.StatusInternalServerError)
		return
	}
	col := rec.Column(indices[0])
	serializedSchema := array.NewStringData(col.Data()).Value(0)
	schema, err := flight.DeserializeSchema([]byte(serializedSchema), memory.DefaultAllocator)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp backend.DataResponse
	resp.Frames = append(resp.Frames, newFrame(schema))
	if err := writeDataResponse(w, resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func newDataResponse(reader recordReader) backend.DataResponse {
	var resp backend.DataResponse
	frame := newFrame(reader.Schema())
READER:
	for reader.Next() {
		record := reader.Record()
		for i, col := range record.Columns() {
			if err := cloneData(frame.Fields[i], col); err != nil {
				resp.Error = err
				break READER
			}
		}
		if err := reader.Err(); err != nil && !errors.Is(err, io.EOF) {
			resp.Error = err
			break
		}
	}
	resp.Frames = append(resp.Frames, frame)
	return resp
}

func writeDataResponse(w io.Writer, resp backend.DataResponse) error {
	json, err := resp.MarshalJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(json)
	return err
}

// NewDatasource creates a new datasource instance
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var cfg config

	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		return nil, fmt.Errorf("FlightSQL Config Unmarshal Error -> %w", err)
	}

	if token, exists := settings.DecryptedSecureJSONData["token"]; exists {
		cfg.Token = token
	}

	if password, exists := settings.DecryptedSecureJSONData["password"]; exists {
		cfg.Password = password
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("FlightSQL Config Validation Error -> ", err)
	}

	client, err := newFlightSQLClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("FlightSQL NewDataSource Error -> %w", err)
	}

	md := createMetadata(cfg)

	if md, err = authenticateClient(ctx, client, cfg, md); err != nil {
		return nil, err
	}

	ds := &DataSource{
		client: client,
		md:     md,
	}
	ds.resourceHandler = route(ds)

	return ds, nil
}

// Dispose cleans up resources before instance is reaped
func (d *DataSource) Dispose() {
	if err := d.client.Close(); err != nil {
		logErrorf(err.Error())
	}
}

// CallResource forwards requests to an internal HTTP mux
func (d *DataSource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return d.resourceHandler.CallResource(ctx, req, sender)
}

// CheckHealth handles health checks sent from Grafana
func (d *DataSource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	query := sqlutil.Query{
		RawSQL: "select 1",
		Format: sqlutil.FormatOptionTable,
	}
	resp := d.query(ctx, query)
	if resp.Error != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("ERROR: %s", resp.Error),
		}, nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "OK",
	}, nil
}

// route sets up the HTTP router with routes and handlers
func route(ds *DataSource) backend.CallResourceHandler {
	r := chi.NewRouter()
	r.Use(recoverer)
	r.Route("/plugin", func(r chi.Router) {
		r.Get("/macros", ds.getMacros)
	})
	r.Route("/flightsql", func(r chi.Router) {
		r.Get("/sql-info", ds.getSQLInfo)
		r.Get("/tables", ds.getTables)
		r.Get("/columns", ds.getColumns)
	})
	return httpadapter.New(r)
}

// createMetadata creates metadata from config
func createMetadata(cfg config) metadata.MD {
	md := metadata.MD{}
	for _, m := range cfg.Metadata {
		for k, v := range m {
			if _, ok := md[k]; !ok && k != "" {
				md.Set(k, v)
			}
		}
	}
	if cfg.Token != "" {
		md.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Token))
	}
	return md
}

// authenticateClient authenticates the client using basic token
func authenticateClient(ctx context.Context, client *client, cfg config, md metadata.MD) (metadata.MD, error) {
	if len(cfg.Username) > 0 || len(cfg.Password) > 0 {
		ctx, err := client.FlightClient().AuthenticateBasicToken(ctx, cfg.Username, cfg.Password)
		if err != nil {
			return nil, fmt.Errorf("FlightSQL Authenticate Error -> %w", err)
		}
		authMD, _ := metadata.FromOutgoingContext(ctx)
		md = metadata.Join(md, authMD)
		return md, nil
	}
	return md, nil
}
