// Package chczap provides log handling using zap package.
// Code structure based on ginrus package.
package chczap

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// Chczap returns a http.Handler (middleware) that logs requests using uber-go/zap.
//
// Requests without errors are logged using zap.Debug().
// TODO: log errors with zap.Error().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
func Chczap(logger *zap.Logger, timeFormat string, utc bool) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			// some evil middlewares modify this values
			path := r.URL.Path
			query := r.URL.RawQuery

			next.ServeHTTP(w, r)

			end := time.Now()
			latency := end.Sub(start)
			if utc {
				end = end.UTC()
			}

			logger.Debug(path,
				zap.Int("status", ww.Status()),
				zap.String("method", r.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", r.RemoteAddr),
				zap.String("user-agent", r.UserAgent()),
				zap.String("time", end.Format(timeFormat)),
				zap.Duration("latency", latency),
			)

		}
		return http.HandlerFunc(fn)
	}
}

// RecoveryWithZap returns a http.Handler (middleware)
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func RecoveryWithZap(logger *zap.Logger, stack bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Check for a broken connection, as it is not really a
					// condition that warrants a panic stack trace.
					var brokenPipe bool
					if ne, ok := err.(*net.OpError); ok {
						if se, ok := ne.Err.(*os.SyscallError); ok {
							if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
								brokenPipe = true
							}
						}
					}

					httpRequest, _ := httputil.DumpRequest(r, false)
					if brokenPipe {
						logger.Error(r.URL.Path,
							zap.Any("error", err),
							zap.String("request", string(httpRequest)),
						)
						// If the connection is dead, we can't write a status to it.
						// TODO: check if context processing is required
						// https://github.com/gin-contrib/zap/blob/270883e70cf28188dc3a8f7e0517fcb3150bd0d8/zap.go#L87-L88
						return
					}

					if stack {
						logger.Error("[Recovery from panic]",
							zap.Time("time", time.Now()),
							zap.Any("error", err),
							zap.String("request", string(httpRequest)),
							zap.String("stack", string(debug.Stack())),
						)
					} else {
						logger.Error("[Recovery from panic]",
							zap.Time("time", time.Now()),
							zap.Any("error", err),
							zap.String("request", string(httpRequest)),
						)
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
