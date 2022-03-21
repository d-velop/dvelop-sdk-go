package otellog_test

import (
	"github.com/d-velop/dvelop-sdk-go/otellog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpGetRequest_LogHttpRequest_AddHttpPropertyAndStatusCodeAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	middleware := otellog.LogHttpRequest()(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	middleware.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "https://www.example.com/path?q=param", nil))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Handle request https://www.example.com/path?q=param with status code '404'\",\"attr\":{\"http\":{\"method\":\"GET\",\"statusCode\":404,\"url\":\"https://www.example.com/path?q=param\",\"target\":\"/path?q=param\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/path\",\"clientIP\":\"192.0.2.1:1234\"}}}\n")
}
