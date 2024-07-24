package main

import (
	"github.com/datalayers-io/grafana-datalayers-datasource/pkg/arrow_flightsql"
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	if err := datasource.Manage("datalayers-arrow-flightsql-datasource", arrow_flightsql.NewDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
