package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/requestid"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.Resolve(r.Header.Get(requestid.Header))
		ctx := requestid.WithContext(r.Context(), id)
		r = r.WithContext(ctx)
		w.Header().Set(requestid.Header, id)

		wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		startedAt := time.Now()
		defer func() {
			if isHealthPath(r.URL.Path) {
				return
			}
			status := wrapped.Status()
			if status == 0 {
				status = http.StatusOK
			}
			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int("response_bytes", wrapped.BytesWritten()),
				slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
			}
			if routePattern := chi.RouteContext(r.Context()).RoutePattern(); routePattern != "" {
				attrs = append(attrs, slog.String("route", routePattern))
			}
			slog.LogAttrs(ctx, httpLogLevel(status), "HTTP-запрос завершён", attrs...)
		}()

		next.ServeHTTP(wrapped, r)
	})
}

func httpLogLevel(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status == http.StatusTooManyRequests:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func isHealthPath(path string) bool {
	return path == "/health/live" || path == "/health/ready"
}
