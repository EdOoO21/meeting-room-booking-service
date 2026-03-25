package http

import (
	stdhttp "net/http"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/ports"
)

type statusRecorder struct {
	stdhttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func requestLogger(logger ports.Logger) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			startedAt := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: stdhttp.StatusOK}

			next.ServeHTTP(recorder, r)

			if logger == nil {
				return
			}

			logger.Info(
				"http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", recorder.status,
				"duration", time.Since(startedAt).String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
