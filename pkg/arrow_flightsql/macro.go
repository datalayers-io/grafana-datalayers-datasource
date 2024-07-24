package arrow_flightsql

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"
)

// Define macros with their corresponding functions.
var macros = sqlutil.Macros{
	"dateBin":        createMacroDateBin(""),
	"dateBinAlias":   createMacroDateBin("_binned"),
	"interval":       macroInterval,
	"timeGroup":      macroTimeGroup,
	"timeGroupAlias": macroTimeGroupAlias,
	"timeRangeFrom":  sqlutil.DefaultMacros["timeFrom"],
	"timeRangeTo":    sqlutil.DefaultMacros["timeTo"],
	"timeRange":      sqlutil.DefaultMacros["timeFilter"],
	"timeTo":         macroTo,
	"timeFrom":       macroFrom,
}

// createMacroDateBin returns a macro function for date_bin.
func createMacroDateBin(suffix string) sqlutil.MacroFunc {
	return func(query *sqlutil.Query, args []string) (string, error) {
		if err := validateArgCount(args, 1); err != nil {
			return "", err
		}
		column := args[0]
		alias := generateAlias(column, suffix)
		return fmt.Sprintf("date_bin(interval '%d second', %s, timestamp '1970-01-01T00:00:00Z')%s", int64(query.Interval.Seconds()), column, alias), nil
	}
}

// macroTimeGroup generates the SQL for time grouping based on the provided arguments.
func macroTimeGroup(query *sqlutil.Query, args []string) (string, error) {
	if err := validateArgCount(args, 2); err != nil {
		return "", err
	}
	column, interval := args[0], args[1]
	return generateTimeGroupSQL(column, interval, false), nil
}

// macroTimeGroupAlias generates the SQL for time grouping with aliases.
func macroTimeGroupAlias(query *sqlutil.Query, args []string) (string, error) {
	if err := validateArgCount(args, 2); err != nil {
		return "", err
	}
	column, interval := args[0], args[1]
	return generateTimeGroupSQL(column, interval, true), nil
}

// macroInterval generates the SQL for interval.
func macroInterval(query *sqlutil.Query, _ []string) (string, error) {
	return fmt.Sprintf("interval '%d second'", int64(query.Interval.Seconds())), nil
}

// macroFrom generates the SQL for the 'from' time range.
func macroFrom(query *sqlutil.Query, _ []string) (string, error) {
	return fmt.Sprintf("cast('%s' as timestamp)", query.TimeRange.From.Format(time.RFC3339)), nil
}

// macroTo generates the SQL for the 'to' time range.
func macroTo(query *sqlutil.Query, _ []string) (string, error) {
	return fmt.Sprintf("cast('%s' as timestamp)", query.TimeRange.To.Format(time.RFC3339)), nil
}

// validateArgCount checks if the number of arguments is as expected.
func validateArgCount(args []string, expected int) error {
	if len(args) != expected {
		return fmt.Errorf("%w: expected %d argument(s), received %d", sqlutil.ErrorBadArgumentCount, expected, len(args))
	}
	return nil
}

// generateAlias creates an alias string based on the column and suffix.
func generateAlias(column, suffix string) string {
	if suffix == "" {
		return ""
	}
	return fmt.Sprintf(" as %s%s", column, suffix)
}

// generateTimeGroupSQL generates the SQL for time grouping.
func generateTimeGroupSQL(column, interval string, withAlias bool) string {
	var res string
	parts := []string{"year", "month", "day", "hour", "minute"}

	for _, part := range parts {
		if interval == part || res != "" {
			alias := ""
			if withAlias {
				alias = fmt.Sprintf(" as %s_%s", column, part)
			}
			if res != "" {
				res += ","
			}
			res += fmt.Sprintf("datepart('%s', %s)%s", part, column, alias)
		}
	}

	return res
}
