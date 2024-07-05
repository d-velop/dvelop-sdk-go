package otellog

import (
	"net/http"
	"strings"
	"time"
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

// LogHttpRequest logs information about the request and response using the provided otel log function
func LogHttpRequest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			before := time.Now()
			WithHttpRequest(r).WithAdditionalAttributes(getAdditionalAttributes(r)).Infof(r.Context(), "########## CALL START: %v - %v", r.Method, r.URL.RequestURI())
			lrw := newLogResponseWriter(w)
			next.ServeHTTP(lrw, r)
			WithHttp(getHttpAttribute(lrw.statusCode, time.Since(before))).WithHttpRequest(r).Infof(r.Context(), "########## CALL FINISH: %v", lrw.statusCode)
		})
	}
}

func getAdditionalAttributes(r *http.Request) map[string]interface{} {
	result := map[string]interface{}{}

	result["headers"] = getHttpHeaders(r)

	return result
}

func getHttpHeaders(r *http.Request) map[string]string {
	resultHeaders := map[string]string{}
	for name, requestHeaders := range r.Header {
		for _, hdr := range requestHeaders {
			if strings.EqualFold(name, "authorization") || strings.EqualFold(name, "cookie") {
				hdr = "***"
			}
			if oldValue, ok := resultHeaders[name]; ok {
				hdr = oldValue + ", " + hdr
			}
			resultHeaders[name] = hdr
		}
	}
	return resultHeaders
}

func getHttpAttribute(statusCode int, elapsed time.Duration) Http {
	return Http{
		StatusCode: uint16(statusCode),
		Server:     &Server{Duration: elapsed},
	}
}
