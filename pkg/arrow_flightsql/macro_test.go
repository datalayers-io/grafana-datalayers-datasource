package arrow_flightsql

import (
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
	"github.com/stretchr/testify/require"
)

func TestMacros(t *testing.T) {
	from := parseTime(t, "2023-01-01T00:00:00Z")

	query := sqlutil.Query{
		TimeRange: backend.TimeRange{
			From: from,
			To:   from.Add(10 * time.Minute),
		},
		Interval: 10 * time.Second,
	}

	testCases := getTestCases()

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			testMacroInterpolation(t, query, tc.in, tc.out)
		})
	}
}

func parseTime(t *testing.T, timeStr string) time.Time {
	t.Helper()
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)
	return parsedTime
}

func getTestCases() []struct {
	in  string
	out string
} {
	return []struct {
		in  string
		out string
	}{
		{
			in:  `select * from x`,
			out: `select * from x`,
		},
		{
			in:  `select date_bin($__interval, time, timestamp '1970-01-01T00:00:00Z')`,
			out: `select date_bin(interval '10 second', time, timestamp '1970-01-01T00:00:00Z')`,
		},
		{
			in:  `select $__dateBin(time)`,
			out: `select date_bin(interval '10 second', time, timestamp '1970-01-01T00:00:00Z')`,
		},
		{
			in:  `select $__dateBinAlias(time)`,
			out: `select date_bin(interval '10 second', time, timestamp '1970-01-01T00:00:00Z') as time_binned`,
		},
		{
			in:  `select * from x where $__timeFilter(time)`,
			out: `select * from x where time >= '2023-01-01T00:00:00Z' AND time <= '2023-01-01T00:10:00Z'`,
		},
		{
			in:  `select * from x where $__timeRangeFrom(time)`,
			out: `select * from x where time >= '2023-01-01T00:00:00Z'`,
		},
		{
			in:  `select * from x where $__timeRangeTo(time)`,
			out: `select * from x where time <= '2023-01-01T00:10:00Z'`,
		},
		{
			in:  `select * from x where $__timeRange(time)`,
			out: `select * from x where time >= '2023-01-01T00:00:00Z' AND time <= '2023-01-01T00:10:00Z'`,
		},
		{
			in:  `select * from x where time >= $__timeFrom`,
			out: `select * from x where time >= cast('2023-01-01T00:00:00Z' as timestamp)`,
		},
		{
			in:  `select * from x where time < $__timeTo`,
			out: `select * from x where time < cast('2023-01-01T00:10:00Z' as timestamp)`,
		},
	}
}

func testMacroInterpolation(t *testing.T, query sqlutil.Query, in, out string) {
	t.Helper()
	sql, err := sqlutil.Interpolate(query.WithSQL(in), macros)
	require.NoError(t, err)
	require.Equal(t, out, sql)
}
