package otellog_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/d-velop/dvelop-sdk-go/otellog"
)

func TestLogHttpRequest_whenServesHttpRequest_thenAddsLogsBeforeAndAfterWithHttpPropertiesAndStatusCodeAndDuration(t *testing.T) {
	logRec := initializeLogger(t)

	req := httptest.NewRequest("GET", "https://www.example.com/path?q=param", nil)
	req.Header["yes"] = []string{"no"}
	req.Header["authorization"] = []string{"my-best-pw"}
	req.Header["multiple"] = []string{"one", "two", "three"}
	req.Header["cookie"] = []string{"some-cookie"}
	rr := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(123)
	})

	m := otellog.LogHttpRequest()
	if m == nil {
		t.Fatal("got nil")
	}

	handler := m(next)
	if handler == nil {
		t.Fatal("middleware did not return a handler")
	}

	handler.ServeHTTP(rr, req)

	output := logRec.String()

	outputWantBeforeDuration := `{"time":"2022-01-01T01:02:03.000000004Z","sev":9,"body":"########## CALL START: GET - /path?q=param","attr":{"headers":{"authorization":"***","cookie":"***","multiple":"one, two, three","yes":"no"},"http":{"clientIP":"192.0.2.1:1234","host":"www.example.com","method":"GET","route":"/path","scheme":"https","target":"/path?q=param","url":"https://www.example.com/path?q=param"}}}
{"time":"2022-01-01T01:02:03.000000004Z","sev":9,"body":"########## CALL FINISH: 123","attr":{"http":{"method":"GET","statusCode":123,"url":"https://www.example.com/path?q=param","target":"/path?q=param","host":"www.example.com","scheme":"https","route":"/path","clientIP":"192.0.2.1:1234","server":{"duration":`
	outputWantAfterDuration := `}}}}` + "\n"

	if !strings.HasPrefix(output, outputWantBeforeDuration) {
		t.Errorf("output does not start with\nExpected: %s\nGot:      %s", outputWantBeforeDuration, output)
	}
	if !strings.HasSuffix(output, outputWantAfterDuration) {
		t.Errorf("output does not end with\nExpected: %s\nGot:      %s", outputWantAfterDuration, output)
	}
}

func TestLogHttpRequest_whenServeHttpRequest_thenCallsNextHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "https://www.example.com/path?q=param", nil)
	rr := httptest.NewRecorder()
	nextWasCalled := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		nextWasCalled = true
	})

	sut := otellog.LogHttpRequest()

	sut(next).ServeHTTP(rr, req)

	if !nextWasCalled {
		t.Error("next http.Handler was not called")
	}
}
