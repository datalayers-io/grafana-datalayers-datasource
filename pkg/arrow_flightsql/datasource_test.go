package arrow_flightsql

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/apache/arrow/go/v12/arrow/flight"
	"github.com/apache/arrow/go/v12/arrow/flight/flightsql"
	"github.com/apache/arrow/go/v12/arrow/flight/flightsql/example"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_QueryData(t *testing.T) {
	db, _, server := setupTestServer(t)
	defer db.Close()
	defer (*server).Shutdown()

	ds := setupDatasource(t, server)
	queryAndVerify(t, ds)
}

func setupTestServer(t *testing.T) (*sql.DB, *example.SQLiteFlightSQLServer, *flight.Server) {
	db, err := example.CreateDB()
	require.NoError(t, err)

	sqliteServer, err := example.NewSQLiteFlightSQLServer(db)
	require.NoError(t, err)
	sqliteServer.Alloc = memory.NewCheckedAllocator(memory.DefaultAllocator)

	server := flight.NewServerWithMiddleware(nil)
	server.RegisterFlightService(flightsql.NewFlightServer(sqliteServer))
	err = server.Init("localhost:0")
	require.NoError(t, err)
	go server.Serve()

	return db, sqliteServer, &server
}

func setupDatasource(t *testing.T, server *flight.Server) *DataSource {
	cfg := config{
		Addr:   (*server).Addr().String(),
		Token:  "secret",
		Secure: false,
	}
	cfgJSON, err := json.Marshal(cfg)
	require.NoError(t, err)

	settings := backend.DataSourceInstanceSettings{JSONData: cfgJSON}
	ds, err := NewDatasource(context.Background(), settings)
	require.NoError(t, err)

	return ds.(*DataSource)
}

func queryAndVerify(t *testing.T, ds *DataSource) {
	resp, err := ds.QueryData(context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{
					RefID: "A",
					JSON:  mustQueryJSON(t, "A", "select * from intTable"),
				},
				{
					RefID: "B",
					JSON:  mustQueryJSON(t, "B", "select 1"),
				},
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, resp.Responses, 2)

	respA := resp.Responses["A"]
	require.NoError(t, respA.Error)
	frame := respA.Frames[0]

	verifyFrame(t, frame)
}

func verifyFrame(t *testing.T, frame *data.Frame) {
	require.Equal(t, "id", frame.Fields[0].Name)
	require.Equal(t, "keyName", frame.Fields[1].Name)
	require.Equal(t, "value", frame.Fields[2].Name)
	require.Equal(t, "foreignId", frame.Fields[3].Name)
	for _, f := range frame.Fields {
		assert.Equal(t, 4, f.Len())
	}
}

func mustQueryJSON(t *testing.T, refID, sql string) []byte {
	t.Helper()

	b, err := json.Marshal(queryRequest{
		RefID:  refID,
		Text:   sql,
		Format: "table",
	})
	if err != nil {
		panic(err)
	}
	return b
}
