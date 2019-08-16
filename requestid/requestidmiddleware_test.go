package requestid_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d-velop/dvelop-sdk-go/requestid"
)

func TestShouldCallInnerHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	innerHandler := handlerSpy{}

	requestid.AddToCtx()(&innerHandler).ServeHTTP(httptest.NewRecorder(), req)

	if !innerHandler.hasBeenCalled {
		t.Error("inner handler should have been called")
	}
}

func TestNoRequestIdHeader_GeneratesNewId(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	innerHandler := handlerSpy{}

	requestid.AddToCtx()(&innerHandler).ServeHTTP(httptest.NewRecorder(), req)

	if err := innerHandler.assertRequestIdIsSet(); err != nil {
		t.Error(err)
	}
}

func TestRequestIdHeader_UsesHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const ReqIdFromHeader = "550e8400-e29b-11d4-a716-446655440000"
	req.Header.Set("x-dv-request-id", ReqIdFromHeader)
	innerHandler := handlerSpy{}

	requestid.AddToCtx()(&innerHandler).ServeHTTP(httptest.NewRecorder(), req)

	if err := innerHandler.assertRequestIdIs(ReqIdFromHeader); err != nil {
		t.Error(err)
	}
}

type handlerSpy struct {
	hasBeenCalled bool
	reqid         string
}

func (spy *handlerSpy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	spy.hasBeenCalled = true
	spy.reqid, _ = requestid.FromCtx(r.Context())
}

func (spy *handlerSpy) assertRequestIdIs(expected string) error {
	if spy.reqid != expected {
		return fmt.Errorf("handler set wrong requestid on context: got %v want %v", spy.reqid, expected)
	}
	return nil
}

func (spy *handlerSpy) assertRequestIdIsSet() error {
	if spy.reqid == "" {
		return fmt.Errorf("handler did not set a requestid on context")
	}
	return nil
}
