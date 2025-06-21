package logger

import (
	"context"
	"github.com/hinha/library-management-synapsis/cmd/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"time"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	w          http.ResponseWriter
	statusCode int
}

// Header returns the header map from the underlying ResponseWriter
func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

// Write writes the data to the underlying ResponseWriter
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.w.Write(b)
}

// WriteHeader captures the status code and writes it to the underlying ResponseWriter
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.w.WriteHeader(statusCode)
}

// NewLogger initializes global zerolog logger and returns:
// - zerolog.Logger instance
// - GORM-compatible logger.Interface
// - gRPC unary interceptor
// - HTTP middleware for logging HTTP requests
func NewLogger() (logger.Interface, grpc.UnaryServerInterceptor, func(http.Handler) http.Handler) {
	// Setup global time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Output to console (pretty)
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	// Set log level
	level := zerolog.InfoLevel
	if config.LogDebug {
		level = zerolog.DebugLevel
	}
	// Global logger with level
	zlogger := zerolog.New(consoleWriter).Level(level).With().Timestamp().Logger()
	log.Logger = zlogger

	// GORM Logger Implementation
	gormLogger := &ZerologGormLogger{
		LogLevel: logger.Info,
		Debug:    config.LogDebug,
	}

	// gRPC Interceptor
	unaryInterceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		event := log.Ctx(ctx).Info()
		resp, err := handler(ctx, req)

		event.
			Str("method", info.FullMethod).
			Dur("duration", time.Since(start)).
			Interface("request", req).
			Err(err).
			Msg("grpc unary")

		return resp, err
	}

	if config.LogDebug {
		log.Debug().Enabled()
	}

	// HTTP Middleware for logging HTTP requests
	httpMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response wrapper to capture status code
			ww := &responseWriter{w: w, statusCode: http.StatusOK}

			// Process the request
			next.ServeHTTP(ww, r)

			// Log the request
			event := log.Info()
			if config.LogDebug {
				event = log.Debug()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", ww.statusCode).
				Dur("duration", time.Since(start)).
				Msg("http request")
		})
	}

	return gormLogger, unaryInterceptor, httpMiddleware
}
