package otellog

import (
	"net/http"
)

type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *logResponseWriter) WriteHeader(status int) {
	r.statusCode = status
	r.ResponseWriter.WriteHeader(status)
}

func newLogResponseWriter(rw http.ResponseWriter) *logResponseWriter {
	return &logResponseWriter{rw, http.StatusOK}
}

func LogHttpRequest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			lrw := newLogResponseWriter(w)
			next.ServeHTTP(lrw, req)
			WithHttpRequest(req).WithHttpStatusCode(lrw.statusCode).Infof(req.Context(), "Handle request %v with status code '%v'", req.URL, lrw.statusCode)
		})
	}
}
