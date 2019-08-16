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
const appNameHeader = "x-dv-signer-name"
const algorithmHeader = "x-dv-signature-algorithm"
const signedHeadersHeader = "x-dv-signed-headers"
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
	req.Header.Set(appNameHeader, "myapp")
	req.Header.Set(signedHeadersHeader, "x-dv-signature-timestamp,x-dv-signer-name,x-dv-signature-algorithm,x-dv-signed-headers")
	req.Header.Set(authorizationHeader, "Bearer 1a6e9cec49889113210d57e9659f0739bb8b1772f16b0ce56792c73d742b991c")
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	timeNow := func() time.Time {
		location, _ := time.LoadLocation("Europe/Berlin")
		return time.Date(2019, time.August, 16, 14, 00, 27, 0, location)
	}

	requestsigner.HandleSignMiddleware([]byte("foobar"), 5, timeNow)(&handlerSpy).ServeHTTP(responseSpy, req)
	if err := responseSpy.assertStatusCodeIs(200); err != nil {
		t.Error(err)
	}
}

// req.method != POST
func TestHandleSignMiddleware_WrongHttpMethodIsUsed_Returns405(t *testing.T) {

}

// req.Header.Accept wrong
func TestHandleSignMiddleware_WrongAcceptHeaderUsed_Returns400(t *testing.T) {

}

// wrong life cylce event path
func TestHandleSignMiddleware_PathIsNotLifeCylceEventPath_Returns400(t *testing.T) {

}

// wrong sign found
func TestHandleSignMiddleware_RequestHasInvalidSignature_Returns403(t *testing.T) {

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

// time stamp is to much in the past

// time stamp is to much in the future

// missing signed headers

// missing headers described by signed headers

// authorization hash not equals calculated hash
