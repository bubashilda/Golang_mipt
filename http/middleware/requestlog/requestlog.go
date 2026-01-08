package requestlog

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := generateRequestID()

			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)

			l.Info("request started",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("request_id", requestID),
			)

			start := time.Now()

			metrics := httpsnoop.CaptureMetricsFn(w, func(ww http.ResponseWriter) {
				defer func() {
					if err := recover(); err != nil {
						l.Error("request panicked",
							zap.String("path", r.URL.Path),
							zap.String("method", r.Method),
							zap.String("request_id", requestID),
							zap.Duration("duration", time.Since(start)),
							zap.Any("panic", err),
						)
						panic(err)
					}
				}()
				next.ServeHTTP(ww, r)
			})

			l.Info("request finished",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("request_id", requestID),
				zap.Duration("duration", metrics.Duration),
				zap.Int("status_code", metrics.Code),
			)
		})
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
