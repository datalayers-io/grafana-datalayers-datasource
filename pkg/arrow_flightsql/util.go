package arrow_flightsql

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// recoverer recovers from a panic and logs the error
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if rec != http.ErrAbortHandler {
					logErrorf("Panic: %s %s", rec, string(debug.Stack()))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func logErrorf(format string, v ...any) {
	log.DefaultLogger.Error(fmt.Sprintf(format, v...))
}
