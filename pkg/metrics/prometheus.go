package metrics

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
	})
}

// StartMetricsServer starts a prometheus server.
// Data Url is at localhost:<port>/metrics/<endpoint>
// Normally you would use /metrics as endpoint and 9090 as port
func StartMetricsServer(endpoint string, port string) chan error {
	router := chi.NewRouter()

	zap.S().Infof("Metrics (prometheus) starting: %v", port)

	// Prometheus endpoint
	router.Get(endpoint, promhttp.Handler().(http.HandlerFunc))
	errChan := make(chan error)

	go func() {
		server := &http.Server{
			Addr:              fmt.Sprintf(":%s", port),
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		}

		err := server.ListenAndServe()
		if err != nil {
			zap.S().Errorf("Prometheus server error: %v", err)
			errChan <- err
		} else {
			zap.S().Infof("Prometheus server serving at port %s", port)
		}
	}()

	return errChan
}
