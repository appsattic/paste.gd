package logger

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/chilts/sid"
	"github.com/gomiddleware/logit"
)

type key int

const logIdKey key = 82

// Logger middleware.
type Logger struct {
	h   http.Handler
	log *logit.Logger
}

// SetLogger sets the logger to `log`. If you have used logger.New(), you can use this to set your
// logger. Alternatively, if you already have your log.Logger, then you can just call logger.NewLogger() directly.
func (l *Logger) SetLogger(log *logit.Logger) {
	l.log = log
}

// wrapper to capture status.
type wrapper struct {
	http.ResponseWriter
	written int
	status  int
}

// capture status.
func (w *wrapper) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// capture written bytes.
func (w *wrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += n
	return n, err
}

// New logger middleware.
func New(sys string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &Logger{
			log: logit.New(os.Stdout, sys),
			h:   h,
		}
	}
}

// NewLogger logger middleware with the given logger.
func NewLogger(log *logit.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &Logger{
			log: log,
			h:   h,
		}
	}
}

// ServeHTTP implementation.
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	res := &wrapper{w, 0, 200}

	// clone the logger and put it into the context
	log := l.log.Clone("req")
	ctx := context.WithValue(r.Context(), logIdKey, log)

	// ToDo: we could use `r.Header.Get("X-Request-ID")` but we should have a config value to determine whether to use
	// it or not, since it could come from outside and be untrusted.

	// set up some fields for this request
	log.WithField("method", r.Method)
	log.WithField("uri", r.RequestURI)
	log.WithField("id", sid.Id())
	log.Print("req-start")

	// continue to the next middleware
	l.h.ServeHTTP(res, r.WithContext(ctx))

	// output the final log line
	log.WithField("status", res.status)
	log.WithField("size", res.written)
	log.WithField("duration", time.Since(start))
	log.Print("req-end")
}

// LogFromRequest can be used to obtain the Log from the request.
func LogFromRequest(r *http.Request) *logit.Logger {
	return r.Context().Value(logIdKey).(*logit.Logger)
}
