package requestsignature_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/d-velop/dvelop-sdk-go/requestsignature"
)

const timestampHeader = "x-dv-signature-timestamp"
const algorithmHeader = "x-dv-signature-algorithm"
const signedHeadersHeader = "x-dv-signature-headers"
const authorizationHeader = "authorization"

const algorithm = "DV1-HMAC-SHA256"

func mockLogInfo(ctx context.Context, msg string) {
	log.Printf("INFO: %v", msg)
}
func mockLogError(ctx context.Context, msg string) {
	log.Printf("ERROR: %v", msg)
}

func TestRequestSigner_ValidateSignedRequest_HappyPath_Working(t *testing.T) {
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}

	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	validator := requestsignature.NewRequestSigner(appSecret, now, mockLogInfo)
	err = validator.ValidateSignedRequest(req)
	if err != nil {
		t.Errorf("got error %v but no error expected", err)
	}
}

func TestRequestSigner_ValidateSignedRequest_AuthorationHeaderInvalid_ReturnError(t *testing.T) {
	wantErrorMessage := "wrong signature in authorization header. Got 12783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104d want 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c"
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}

	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 12783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104d",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	validator := requestsignature.NewRequestSigner(appSecret, now, mockLogInfo)
	err = validator.ValidateSignedRequest(req)
	if err != nil {
		if err.Error() != wantErrorMessage {
			t.Errorf("wrong error returned: got %v want %v", err, wantErrorMessage)
		}
	} else {
		t.Errorf("no error returned, but want error %v", wantErrorMessage)
	}
}

func TestRequestSigner_ValidateSignedRequest_WithWrongDto_ReturnError(t *testing.T) {
	wantErrorMessage := "wrong signature in authorization header. Got 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c want daba3a1deb11b646540bcb42161ea0003cf6ca6c1c3282d83e8c80e91cfcd9f9"
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}

	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://xyz.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	validator := requestsignature.NewRequestSigner(appSecret, now, mockLogInfo)
	err = validator.ValidateSignedRequest(req)
	if err != nil {
		if err.Error() != wantErrorMessage {
			t.Errorf("wrong error returned: got '%v' want '%v'", err, wantErrorMessage)
		}
	} else {
		t.Errorf("no error returned, but want error %v", wantErrorMessage)
	}
}

func TestRequestSigner_ValidateSignedRequest_RequestTimeouted_ReturnError(t *testing.T) {
	wantErrorMessage := "request is timed out: timestamp from request: 2019-08-09T08:49:42Z, current time: 2019-08-09T09:49:45Z"
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}

	now := func() time.Time {
		return time.Date(2019, time.August, 9, 9, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	validator := requestsignature.NewRequestSigner(appSecret, now, mockLogInfo)
	err = validator.ValidateSignedRequest(req)
	if err != nil {
		if err.Error() != wantErrorMessage {
			t.Errorf("wrong error returned: got %v want %v", err, wantErrorMessage)
		}
	} else {
		t.Errorf("no error returned, but want error %v", wantErrorMessage)
	}
}

func TestHandleSignMiddleware_HappyPath_Working(t *testing.T) {
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}

	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	handlerCalled := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		t.Log("handler was called")
	}

	rr := httptest.NewRecorder()
	requestsignature.HandleCloudSignatureMiddleware(appSecret, now, mockLogInfo, mockLogError)(http.HandlerFunc(handler)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status code returned: got %v want %v", status, http.StatusOK)
	}
	if !handlerCalled {
		t.Fatalf("test handler not called")
	}
}

func TestHandleSignMiddleware_AppSecretMissing_Return500InternalServerError(t *testing.T) {
	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	handlerCalled := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		t.Log("handler was called")
	}

	rr := httptest.NewRecorder()
	requestsignature.HandleCloudSignatureMiddleware(nil, now, mockLogInfo, mockLogError)(http.HandlerFunc(handler)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("wrong status code returned: got %v want %v", status, http.StatusInternalServerError)
	}
	if handlerCalled {
		t.Fatalf("Handler was called, but should not be called.")
	}
}

func TestHandleSignMiddleware_MiddlewareWasCalledByWrongMethod_Return405MethodNotAllowed(t *testing.T) {
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}
	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodGet, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	handlerCalled := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		t.Log("handler was called")
	}

	rr := httptest.NewRecorder()
	requestsignature.HandleCloudSignatureMiddleware(appSecret, now, mockLogInfo, mockLogError)(http.HandlerFunc(handler)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Fatalf("wrong status code returned: got %v want %v", status, http.StatusMethodNotAllowed)
	}
	if handlerCalled {
		t.Fatalf("Handler was called, but should not be called.")
	}
}

func TestHandleSignMiddleware_MiddlewareWasCalledByPath_Return400BadRequest(t *testing.T) {
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}
	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event/wrongpath", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	handlerCalled := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		t.Log("handler was called")
	}

	rr := httptest.NewRecorder()
	requestsignature.HandleCloudSignatureMiddleware(appSecret, now, mockLogInfo, mockLogError)(http.HandlerFunc(handler)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("wrong status code returned: got %v want %v", status, http.StatusBadRequest)
	}
	if handlerCalled {
		t.Fatalf("Handler was called, but should not be called.")
	}
}

func TestHandleSignMiddleware_MiddlewareWasCalledWithoutContentTypeHeader_Return406NotAcceptable(t *testing.T) {
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}
	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"id",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		//"Content-Type": 			"application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	handlerCalled := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		t.Log("handler was called")
	}

	rr := httptest.NewRecorder()
	requestsignature.HandleCloudSignatureMiddleware(appSecret, now, mockLogInfo, mockLogError)(http.HandlerFunc(handler)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotAcceptable {
		t.Fatalf("wrong status code returned: got %v want %v", status, http.StatusBadRequest)
	}
	if handlerCalled {
		t.Fatalf("Handler was called, but should not be called.")
	}
}

func TestHandleSignMiddleware_MiddlewareWasCalledButSignatureIsInvalid_Return403Forbidden(t *testing.T) {
	appSecret, err := base64.StdEncoding.DecodeString("Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=")
	if err != nil {
		t.Fatalf("app secret string is not valid base64 encoded string. Error = %v", err)
	}
	now := func() time.Time {
		return time.Date(2019, time.August, 9, 8, 49, 45, 0, time.UTC)
	}

	dto := requestsignature.Dto{
		"subscribe",
		"wrong id to generate wrong body hash",
		"https://someone.d-velop.cloud",
	}

	headers := map[string]string{
		"x-dv-signature-headers":   "x-dv-signature-algorithm,x-dv-signature-headers,x-dv-signature-timestamp",
		"x-dv-signature-algorithm": "DV1-HMAC-SHA256",
		"x-dv-signature-timestamp": "2019-08-09T08:49:42Z",
		"Authorization":            "Bearer 02783453441665bf27aa465cbbac9b98507ae94c54b6be2b1882fe9a05ec104c",
		"Content-Type":             "application/json",
	}

	payload := &bytes.Buffer{}
	json.NewEncoder(payload).Encode(dto)
	body := payload.Bytes()
	req, _ := http.NewRequest(http.MethodPost, "/myapp/dvelop-cloud-lifecycle-event", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	handlerCalled := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		t.Log("handler was called")
	}

	rr := httptest.NewRecorder()
	requestsignature.HandleCloudSignatureMiddleware(appSecret, now, mockLogInfo, mockLogError)(http.HandlerFunc(handler)).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusForbidden {
		t.Fatalf("wrong status code returned: got %v want %v", status, http.StatusForbidden)
	}
	if handlerCalled {
		t.Fatalf("Handler was called, but should not be called.")
	}
}
