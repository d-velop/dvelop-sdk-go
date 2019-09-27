package tenant_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d-velop/dvelop-sdk-go/tenant"

	"encoding/base64"

	"crypto/hmac"
	"crypto/sha256"
)

const systemBaseUriHeader = "x-dv-baseuri"
const tenantIdHeader = "x-dv-tenant-id"
const signatureHeader = "x-dv-sig-1"
const defaultSystemBaseUri = "https://default.example.com"

func TestBaseUriHeaderAndEmptyDefaultBaseUri_UsesHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const systemBaseUriFromHeader = "https://sample.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	req.Header.Set(signatureHeader, base64Signature(systemBaseUriFromHeader, signatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertBaseUriIs(systemBaseUriFromHeader); err != nil {
		t.Error(err)
	}
}

func TestNoBaseUriHeaderAndDefaultBaseUri_UsesDefaultBaseUri(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx(defaultSystemBaseUri, signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertBaseUriIs(defaultSystemBaseUri); err != nil {
		t.Error(err)
	}
}

func TestBaseUriHeaderAndDefaultBaseUri_UsesHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const systemBaseUriFromHeader = "https://header.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	req.Header.Set(signatureHeader, base64Signature(systemBaseUriFromHeader, signatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx(defaultSystemBaseUri, signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertBaseUriIs(systemBaseUriFromHeader); err != nil {
		t.Error(err)
	}
}

func TestNoBaseUriHeaderAndEmptyDefaultBaseUri_DoesntAddBaseUriToContext(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	handlerSpy := handlerSpy{}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertErrorReadingSystemBaseUri(); err != nil {
		t.Error(err)
	}
}

func TestTenantIdHeader_UsesHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	req.Header.Set(signatureHeader, base64Signature(tenantIdFromHeader, signatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertTenantIdIs(tenantIdFromHeader); err != nil {
		t.Error(err)
	}
}

func TestNoTenantIdHeader_UsesTenantIdZero(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertTenantIdIs("0"); err != nil {
		t.Error(err)
	}
}

func TestTenantIdHeaderAndBaseUriHeader_UsesHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	const systemBaseUriFromHeader = "https://header.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	req.Header.Set(signatureHeader, base64Signature(systemBaseUriFromHeader+tenantIdFromHeader, signatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx(defaultSystemBaseUri, signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertTenantIdIs(tenantIdFromHeader); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertBaseUriIs(systemBaseUriFromHeader); err != nil {
		t.Error(err)
	}
}

func TestTenantIdHeaderAndNoBaseUriHeader_UsesTenantIdHeaderAndDefaultSystemBaseUri(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	req.Header.Set(signatureHeader, base64Signature(tenantIdFromHeader, signatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx(defaultSystemBaseUri, signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertTenantIdIs(tenantIdFromHeader); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertBaseUriIs(defaultSystemBaseUri); err != nil {
		t.Error(err)
	}
}

func TestNoHeadersButDefaultSystemBaseUri_UsesDefaultBaseUriAndTenantIdZero(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx(defaultSystemBaseUri, signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertTenantIdIs("0"); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertBaseUriIs(defaultSystemBaseUri); err != nil {
		t.Error(err)
	}
}

func TestNoHeadersButDefaultSystemBaseUriAndNoSignatureSecretKey_UsesDefaultBaseUriAndTenantIdZero(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx(defaultSystemBaseUri, nil)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusOK); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertTenantIdIs("0"); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertBaseUriIs(defaultSystemBaseUri); err != nil {
		t.Error(err)
	}
}

func TestWrongDataSignedWithValidSignatureKey_Returns403(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const systemBaseUriFromHeader = "https://sample.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	req.Header.Set(signatureHeader, base64Signature("wrong data", signatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusForbidden); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestNoneBase64Signature_Returns403(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const systemBaseUriFromHeader = "https://sample.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	req.Header.Set(signatureHeader, "abc+(9-!")
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusForbidden); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestWrongSignatureKey_Returns403(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const systemBaseUriFromHeader = "https://sample.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	wrongSignatureKey := []byte{167, 219, 144, 209, 189, 1, 178, 73, 139, 47, 21, 236, 142, 56, 71, 245, 43, 188, 163, 52, 239, 102, 94, 153, 255, 159, 199, 149, 163, 145, 161, 24}
	req.Header.Set(signatureHeader, base64Signature(systemBaseUriFromHeader+tenantIdFromHeader, wrongSignatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusForbidden); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestHeadersWithoutSignature_Returns403(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const systemBaseUriFromHeader = "https://sample.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", signatureKey)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusForbidden); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestHeadersAndNoSignatureSecretKey_Returns500(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	const systemBaseUriFromHeader = "https://sample.example.com"
	req.Header.Set(systemBaseUriHeader, systemBaseUriFromHeader)
	const tenantIdFromHeader = "a12be5"
	req.Header.Set(tenantIdHeader, tenantIdFromHeader)
	req.Header.Set(signatureHeader, base64Signature(systemBaseUriFromHeader+tenantIdFromHeader, signatureKey))
	handlerSpy := handlerSpy{}
	responseSpy := responseSpy{httptest.NewRecorder()}

	tenant.AddToCtx("", nil)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestNoIdOnContext_SetId_ReturnsContextWithId(t *testing.T) {
	ctx := tenant.SetId(context.Background(), "123ABC")
	if id, _ := tenant.IdFromCtx(ctx); id != "123ABC" {
		t.Errorf("got wrong tenantId from context: got %v want %v", id, "123ABC")
	}
}

func TestIdOnContext_SetId_ReturnsContextWithNewId(t *testing.T) {
	ctx := tenant.SetId(context.Background(), "123ABC")
	ctx = tenant.SetId(ctx, "XYZ")
	if id, _ := tenant.IdFromCtx(ctx); id != "XYZ" {
		t.Errorf("got wrong tenantId from context: got %v want %v", id, "XYZ")
	}
}

func TestSystemBaseUriOnContext_SetSystemBaseUri_ReturnsContextWithSystemBaseUri(t *testing.T) {
	ctx := tenant.SetSystemBaseUri(context.Background(), "https://xyz.example.com")
	if u, _ := tenant.SystemBaseUriFromCtx(ctx); u != "https://xyz.example.com" {
		t.Errorf("got wrong systemBaseUri from context: got %v want %v", u, "https://xyz.example.com")
	}
}

func TestSystemBaseUriOnContext_SetSystemBaseUri_ReturnsContextWithNewSystemBaseUri(t *testing.T) {
	ctx := tenant.SetSystemBaseUri(context.Background(), "https://xyz.example.com")
	ctx = tenant.SetSystemBaseUri(context.Background(), "https://abc.example.com")
	if u, _ := tenant.SystemBaseUriFromCtx(ctx); u != "https://abc.example.com" {
		t.Errorf("got wrong systemBaseUri from context: got %v want %v", u, "https://abc.example.com")
	}
}

var signatureKey = []byte{166, 219, 144, 209, 189, 1, 178, 73, 139, 47, 21, 236, 142, 56, 71, 245, 43, 188, 163, 52, 239, 102, 94, 153, 255, 159, 199, 149, 163, 145, 161, 24}

func base64Signature(message string, sigKey []byte) string {
	mac := hmac.New(sha256.New, sigKey)
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

type handlerSpy struct {
	systemBaseUri             string
	tenantId                  string
	errorReadingSystemBaseUri error
	errorReadingTenantId      error
	hasBeenCalled             bool
}

func (spy *handlerSpy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	spy.hasBeenCalled = true
	spy.systemBaseUri, spy.errorReadingSystemBaseUri = tenant.SystemBaseUriFromCtx(r.Context())
	spy.tenantId, spy.errorReadingTenantId = tenant.IdFromCtx(r.Context())
}

func (spy *handlerSpy) assertBaseUriIs(expected string) error {
	if spy.systemBaseUri != expected {
		return fmt.Errorf("handler set wrong systemBaseUri on context: got %v want %v", spy.systemBaseUri, expected)
	}
	return nil
}

func (spy *handlerSpy) assertTenantIdIs(expected string) error {
	if spy.tenantId != expected {
		return fmt.Errorf("handler set wrong tenantId on context: got %v want %v", spy.tenantId, expected)
	}
	return nil
}

func (spy *handlerSpy) assertErrorReadingSystemBaseUri() error {
	if spy.errorReadingSystemBaseUri == nil {
		return fmt.Errorf("expected error while reading systemBaseUri from context")
	}
	return nil
}

func (spy *handlerSpy) assertErrorReadingTenantId() error {
	if spy.errorReadingTenantId == nil {
		return fmt.Errorf("expected error while reading tenantId from context")
	}
	return nil
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
