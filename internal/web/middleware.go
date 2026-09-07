package web

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status before passing it through. If a handler never
// calls WriteHeader, the recorder keeps the 200 default set by the middleware.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (app *App) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Observe the response status without reading or copying the request body.
		recorder := statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(&recorder, r)

		// Keep logs to routing metadata only; user text, substrings, cookies,
		// headers, and generated output are deliberately excluded.
		app.logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", recorder.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
