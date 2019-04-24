package requestlog_test

import (
	"context"
	"fmt"
	"github.com/d-velop/dvelop-sdk-go/requestlog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestShouldCallInnerHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	innerHandler := handlerMock{}

	requestlog.Log(func(ctx context.Context, logmessage string) {
	})(&innerHandler).ServeHTTP(httptest.NewRecorder(), req)

	if !innerHandler.hasBeenCalled {
		t.Error("inner handler should have been called")
	}
}

func TestShouldCallLogFunction(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	logfunctionHasBeenCalled := false

	requestlog.Log(func(ctx context.Context, logmessage string) {
		logfunctionHasBeenCalled = true
	})(&handlerMock{}).ServeHTTP(httptest.NewRecorder(), req)

	if !logfunctionHasBeenCalled {
		t.Error("log function should have been called")
	}
}

func TestShouldLogBeginOfRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Accept", "text/html")
	loggedMessages := make([]string, 0)

	requestlog.Log(func(ctx context.Context, logmessage string) {
		loggedMessages = append(loggedMessages, logmessage)
	})(&handlerMock{}).ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(loggedMessages[0], "GET") {
		t.Errorf("Logmessage '%v' should contain Http Method 'GET'", loggedMessages[0])
	}
	if !strings.Contains(loggedMessages[0], "/myresource/subs") {
		t.Errorf("Logmessage '%v' should contain URL '/myresource/sub'", loggedMessages[0])
	}
	if !strings.Contains(loggedMessages[0], "Accept") || !strings.Contains(loggedMessages[0], "text/html") {
		t.Errorf("Logmessage '%v' should contain request header '%v' with value '%v'", loggedMessages[0], "Accept", "text/html")
	}
}

func TestShouldLogEndOfRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	loggedMessages := make([]string, 0)

	requestlog.Log(func(ctx context.Context, logmessage string) {
		loggedMessages = append(loggedMessages, logmessage)
	})(&handlerMock{}).ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(loggedMessages[1], "GET") {
		t.Errorf("Logmessage '%v' should contain Http Method 'GET'", loggedMessages[1])
	}
	if !strings.Contains(loggedMessages[1], "/myresource/sub") {
		t.Errorf("Logmessage '%v' should contain URL '/myresource/sub'", loggedMessages[1])
	}
	elapsedTimeRegex := regexp.MustCompile(`millis="\d*"`)
	if !elapsedTimeRegex.MatchString(loggedMessages[1]) {
		t.Errorf("Logmessage '%v' should contain elapsed time in millis", loggedMessages[1])
	}
	if !strings.Contains(loggedMessages[1], "status=\"202\"") {
		t.Errorf("Logmessage '%v' should contain status '202'", loggedMessages[1])
	}
	if !strings.Contains(loggedMessages[1], "Content-Type") || !strings.Contains(loggedMessages[1], "text/html; charset=utf-8") {
		t.Errorf("Logmessage '%v' should contain response header '%v' with value '%v'", loggedMessages[1], "Content-Type", "text/html; charset=utf-8")
	}
}

func TestShouldNotLogSensitiveResponseData(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	loggedMessages := make([]string, 0)

	requestlog.Log(func(ctx context.Context, logmessage string) {
		loggedMessages = append(loggedMessages, logmessage)
	})(&handlerMock{}).ServeHTTP(httptest.NewRecorder(), req)

	if strings.Contains(loggedMessages[1], "a%sb125%sfaffqwERFW13-ads23") {
		t.Errorf("Logmessage '%v' should NOT contain AuthSessionId '%v'", loggedMessages[1], "a%sb125%sfaffqwERFW13-ads23")
	}
}

func TestShouldNotLogSensitiveRequestData(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", "hubspotutk=74h2cf56; AuthSessionId=a%sb125%sfaffqwERFW13-ads23; rID=c10161a5-8724-484b-802e-abaa51a6779e")
	req.Header.Set("Authorization", "Basic YWxhZGRpbjpvcGVuc2VzYW1l")
	loggedMessages := make([]string, 0)

	requestlog.Log(func(ctx context.Context, logmessage string) {
		loggedMessages = append(loggedMessages, logmessage)
	})(&handlerMock{}).ServeHTTP(httptest.NewRecorder(), req)

	if strings.Contains(loggedMessages[0], "a%sb125%sfaffqwERFW13-ads23") {
		t.Errorf("Logmessage '%v' should NOT contain AuthSessionId '%v'", loggedMessages[0], "a%sb125%sfaffqwERFW13-ads23")
	}
	if strings.Contains(loggedMessages[0], "YWxhZGRpbjpvcGVuc2VzYW1l") {
		t.Errorf("Logmessage '%v' should NOT contain Authorization credentials '%v'", loggedMessages[0], "YWxhZGRpbjpvcGVuc2VzYW1l")
	}
	if !strings.Contains(loggedMessages[0], "Cookie") || !strings.Contains(loggedMessages[0], "hubspotutk=74h2cf56") || !strings.Contains(loggedMessages[0], "rID=c10161a5-8724-484b-802e-abaa51a6779e") {
		t.Errorf("Logmessage '%v' should contain request header '%v' with values '%v'", loggedMessages[0], "Cookie", [...]string{"hubspotutk=74h2cf56", "rID=c10161a5-8724-484b-802e-abaa51a6779e"})
	}
}

func TestShouldCreateProperHttpResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	requestlog.Log(func(ctx context.Context, logmessage string) {
	})(&handlerMock{}).ServeHTTP(recorder, req)

	if recorder.Body.String() != "Result" {
		t.Errorf("Response body should be '%v' but was '%v'", "Result", recorder.Body.String())
	}
	if recorder.Code != 202 {
		t.Errorf("Response code should be '%v' but was '%v'", "202", recorder.Code)
	}
	if recorder.Result().Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("'Content-Type' response header should be '%v' but was '%v'", "text/html; charset=utf-8", recorder.Result().Header.Get("Content-Type"))
	}
}

type handlerMock struct {
	hasBeenCalled bool
}

func (spy *handlerMock) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	spy.hasBeenCalled = true
	rw.Header().Set("Set-Cookie", "AuthSessionId=a%sb125%sfaffqwERFW13-ads23; path=/; secure; httponly")
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprint(rw, "Result")
}
