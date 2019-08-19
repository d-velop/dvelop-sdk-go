package requestsigner_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/d-velop/dvelop-sdk-go/requestsigner"
)

const timestampHeader = "x-dv-signature-timestamp"
const algorithmHeader = "x-dv-signature-algorithm"
const signedHeadersHeader = "x-dv-signature-headers"
const authorizationHeader = "authorization"

const algorithm = "DV1-HMAC-SHA256"

func TestHandleSignMiddleware_HappyPath_Works(t *testing.T) {
	body := []byte(`{"type":"subscribe","tenantId":"vw","baseUri":"https://myfancy.d-velop.cloud"}`)
	req, err := http.NewRequest("POST", "https://myapp.service.d-velop.cloud/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set(timestampHeader, "2019-08-16T12:00:27Z")
	req.Header.Set(algorithmHeader, algorithm)
	req.Header.Set(signedHeadersHeader, "x-dv-signature-timestamp,x-dv-signature-algorithm,x-dv-signature-headers")
	req.Header.Set(authorizationHeader, "Bearer 1ed80bcfdfdd67455413fa4a2b47e013aca9f1a846501f545931bead0e21cb12")
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	timeNow := func() time.Time {
		location, _ := time.LoadLocation("Europe/Berlin")
		return time.Date(2019, time.August, 16, 14, 00, 27, 0, location)
	}

	requestsigner.HandleSignMiddleware([]byte("foobar"), timeNow)(&handlerSpy).ServeHTTP(responseSpy, req)
	if err := responseSpy.assertStatusCodeIs(200); err != nil {
		t.Error(err)
	}
}

func TestHandleSignMiddleware_AppSecretNotSet_Returns500(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://foobar.com", nil)
	h := handlerSpy{}
	w := responseSpy{httptest.NewRecorder()}
	requestsigner.HandleSignMiddleware(nil, nil)(&h).ServeHTTP(w, req)
	if err := w.assertStatusCodeIs(500); err != nil {
		t.Error(err)
	}
}

// req.method != POST
func TestHandleSignMiddleware_WrongHttpMethodIsUsed_Returns405(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://foobar.com", nil)
	h := handlerSpy{}
	w := responseSpy{httptest.NewRecorder()}
	requestsigner.HandleSignMiddleware([]byte("foobar"), nil)(&h).ServeHTTP(w, req)
	if err := w.assertStatusCodeIs(405); err != nil {
		t.Error(err)
	}
}

// wrong life cylce event path
func TestHandleSignMiddleware_PathIsNotLifeCylceEventPath_Returns400(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", nil)
	h := handlerSpy{}
	w := responseSpy{httptest.NewRecorder()}
	requestsigner.HandleSignMiddleware([]byte("foobar"), nil)(&h).ServeHTTP(w, req)
	if err := w.assertStatusCodeIs(400); err != nil {
		t.Error(err)
	}
}

// req.Header.Accept wrong
func TestHandleSignMiddleware_WrongAcceptHeaderUsed_Returns400(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar/dvelop-cloud-lifecycle-event", nil)
	h := handlerSpy{}
	w := responseSpy{httptest.NewRecorder()}
	requestsigner.HandleSignMiddleware([]byte("foobar"), nil)(&h).ServeHTTP(w, req)
	if err := w.assertStatusCodeIs(400); err != nil {
		t.Error(err)
	}
}

// wrong sign found
func TestHandleSignMiddleware_RequestHasInvalidSignature_Returns403(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar/dvelop-cloud-lifecycle-event", nil)
	req.Header.Set("accept", "application/json")
	h := handlerSpy{}
	w := responseSpy{httptest.NewRecorder()}
	requestsigner.HandleSignMiddleware([]byte("foobar"), time.Now)(&h).ServeHTTP(w, req)
	if err := w.assertStatusCodeIs(403); err != nil {
		t.Error(err)
	}
}

type handlerSpy struct {
	hasBeenCalled bool
}

func (spy *handlerSpy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	spy.hasBeenCalled = true
}

type responseSpy struct {
	*httptest.ResponseRecorder
}

func (spy *responseSpy) assertStatusCodeIs(expectedStatusCode int) error {
	if status := spy.Code; status != expectedStatusCode {
		return fmt.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatusCode)
	}
	return nil
}

// working
func TestRequestSigner_ValidateSignedRequest_HappyPath_Works(t *testing.T) {
	getNow := func() time.Time {
		location, _ := time.LoadLocation("Europe/Berlin")
		return time.Date(2019, time.August, 5, 13, 39, 27, 4711, location)
	}
	dto := `{"type": "subscribe","tenantId":"vw","baseUri":"https://mycloud.d-velop.cloud"}`
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", bytes.NewBuffer([]byte(dto)))
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer 997bb670e6c61ed1de186c68d7016e01a67045353be9908fa5c53df61507559c")
	req.Header.Set("x-dv-signature-timestamp", "2019-08-05T11:39:27Z")
	req.Header.Set("x-dv-signature-headers", "x-dv-signuature-headers, x-dv-signature-algorithm, x-dv-signature-timestamp")
	req.Header.Set("x-dv-signature-algorithm", "DV1-HMAC-SHA256")

	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), getNow)
	err := requestSigner.ValidateSignedRequest(req)
	if err != nil {
		t.Errorf("no error expected but got error %v", err)
	}
}

func TestRequestSigner_ValidateSignedRequest_AppSecretNotConfigured_ReturnsError(t *testing.T) {
	requestSigner := requestsigner.NewRequestSigner(nil, nil)
	err := requestSigner.ValidateSignedRequest(nil)
	if err == nil {
		t.Errorf("error expected but no error returned")
	}
	if expectedError := "app secret has not been configured"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}

// invalid accept header
func TestRequestSigner_ValidateSignedRequest_WrongAcceptHeaderInRequest_ReturnsError(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", nil)
	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), nil)
	err := requestSigner.ValidateSignedRequest(req)
	if err == nil {
		t.Errorf("error expected but no error returned")
	}
	if expectedError := "wrong accept header found. Got  want application/json"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}

// authorization header missing on request
func TestRequestSigner_ValidateSignedRequest_AuthorizationHeaderMissingInRequest_ReturnsError(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", nil)
	req.Header.Set("accept", "application/json")
	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), nil)
	err := requestSigner.ValidateSignedRequest(req)
	if err == nil {
		t.Errorf("error expected but no error returned")
	}
	if expectedError := "authorization header missing"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}

// invalid bearer token found
func TestRequestSigner_ValidateSignedRequest_AuthorizationBearerTokenInvalid_ReturnsError(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", nil)
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer foobar")
	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), nil)
	err := requestSigner.ValidateSignedRequest(req)
	if err == nil {
		t.Errorf("error expected but no error returned")
	}
	if expectedError := "found authorization header is not a valid Bearer token. Got Bearer foobar"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}

// time stamp is to much in the past
func TestRequestSigner_ValidateSignedRequest_TimestampIs10MinutesInThePast_ReturnsError(t *testing.T) {
	getNow := func() time.Time {
		location, _ := time.LoadLocation("Europe/Berlin")
		return time.Date(2019, time.August, 5, 13, 39, 27, 4711, location)
	}

	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", nil)
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer 0adf")
	req.Header.Set("x-dv-signature-timestamp", "2019-08-05T11:29:27Z")

	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), getNow)
	err := requestSigner.ValidateSignedRequest(req)
	if err == nil {
		t.Errorf("error expected but no error returned")
	}
	if expectedError := "request is timed out: timestamp from request: 2019-08-05T11:29:27Z, current time: 2019-08-05T11:39:27Z"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}

// time stamp is to much in the future
func TestRequestSigner_ValidateSignedRequest_TimestampIs10MinutesInTheFuture_ReturnsError(t *testing.T) {
	getNow := func() time.Time {
		location, _ := time.LoadLocation("Europe/Berlin")
		return time.Date(2019, time.August, 5, 13, 39, 27, 4711, location)
	}

	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", nil)
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer 0adf")
	req.Header.Set("x-dv-signature-timestamp", "2019-08-05T11:49:27Z")

	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), getNow)
	err := requestSigner.ValidateSignedRequest(req)
	if err == nil {
		t.Errorf("error expected but no error returned")
	}
	if expectedError := "request is timed out: timestamp from request: 2019-08-05T11:49:27Z, current time: 2019-08-05T11:39:27Z"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}

// payload missing
func TestRequestSigner_ValidateSignedRequest_PayloadMissingInRequest_ReturnsError(t *testing.T) {
	getNow := func() time.Time {
		location, _ := time.LoadLocation("Europe/Berlin")
		return time.Date(2019, time.August, 5, 13, 39, 27, 4711, location)
	}

	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", nil)
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer 0adf")
	req.Header.Set("x-dv-signature-timestamp", "2019-08-05T11:39:27Z")

	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), getNow)
	err := requestSigner.ValidateSignedRequest(req)
	if err == nil {
		t.Errorf("error expected but no error returned")
	}
	if expectedError := "payload missing"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}

// authorization hash not equals calculated hash
func TestRequestSigner_ValidateSignedRequest_InvalidAuthorizationBearerToken_ReturnsError(t *testing.T) {
	getNow := func() time.Time {
		location, _ := time.LoadLocation("Europe/Berlin")
		return time.Date(2019, time.August, 5, 13, 39, 27, 4711, location)
	}
	dto := `{"type": "subscribe","tenantId":"vw","baseUri":"https://mycloud.d-velop.cloud"}`
	req, _ := http.NewRequest("POST", "https://foobar.com/foo/bar", bytes.NewBuffer([]byte(dto)))
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer 0adf")
	req.Header.Set("x-dv-signature-timestamp", "2019-08-05T11:39:27Z")
	req.Header.Set("x-dv-signature-headers", "x-dv-signuature-headers, x-dv-signature-algorithm, x-dv-signature-timestamp")
	req.Header.Set("x-dv-signature-algorithm", "DV1-HMAC-SHA256")

	requestSigner := requestsigner.NewRequestSigner([]byte("foobar"), getNow)
	err := requestSigner.ValidateSignedRequest(req)
	if err == nil {
		t.Errorf("error expected but no error returned")
		return
	}
	if expectedError := "wrong authorization header. Got 0adf want 997bb670e6c61ed1de186c68d7016e01a67045353be9908fa5c53df61507559c"; err.Error() != expectedError {
		t.Errorf("wrong error returned: got %v want %v", err, expectedError)
	}
}
