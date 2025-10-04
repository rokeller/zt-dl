package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type logMiddlewareWriter struct {
	http.ResponseWriter

	status int
}

func (w *logMiddlewareWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *logMiddlewareWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

type logMiddleware struct {
}

func NewLogMiddleware() *logMiddleware {
	return &logMiddleware{}
}

func (m *logMiddleware) Func() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			ww := &logMiddlewareWriter{ResponseWriter: w}
			next.ServeHTTP(ww, r)

			d := time.Since(startTime)
			fmt.Printf("method=%s  uri=%s  status:%d duration:%s\n",
				r.Method, r.RequestURI, ww.status, d)
		})
	}
}
