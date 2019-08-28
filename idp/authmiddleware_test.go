package idp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/d-velop/dvelop-sdk-go/idp"
	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

func TestGetRequestWithFalseAuthorizationType_RedirectsToIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseSpy := responseSpy{httptest.NewRecorder()}
	req.Header.Set("Authorization", "Basic adadbk")
	handlerSpy := handlerSpy{}

	idp.HandleAuth(nil, nil, false, log, log)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusFound); err != nil {
		t.Error(err)
	}
	if err := responseSpy.assertHeadersAre(map[string]string{"Location": "/identityprovider/login?redirect=%2Fmyresource%2Fsubresource%3Fquery1%3Dabc%26query2%3D123"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestGetRequestWithFalseCookie_RedirectsToIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseSpy := responseSpy{httptest.NewRecorder()}
	req.AddCookie(&http.Cookie{Name: "AnyCookie", Value: "adadbk"})
	handlerSpy := handlerSpy{}

	idp.HandleAuth(nil, nil, false, log, log)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusFound); err != nil {
		t.Error(err)
	}
	if err := responseSpy.assertHeadersAre(map[string]string{"Location": "/identityprovider/login?redirect=%2Fmyresource%2Fsubresource%3Fquery1%3Dabc%26query2%3D123"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestHeadRequestWithoutAuthorizationInfos_RedirectsToIdp(t *testing.T) {
	req, err := http.NewRequest("HEAD", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseSpy := responseSpy{httptest.NewRecorder()}
	handlerSpy := &handlerSpy{}
	idp.HandleAuth(nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusFound); err != nil {
		t.Error(err)
	}
	if err := responseSpy.assertHeadersAre(map[string]string{"Location": "/identityprovider/login?redirect=%2Fmyresource%2Fsubresource%3Fquery1%3Dabc%26query2%3D123"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestPostRequestWithoutAuthorizationInfos_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("POST", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseSpy := responseSpy{httptest.NewRecorder()}
	handlerSpy := &handlerSpy{}
	idp.HandleAuth(nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if err := responseSpy.assertHeadersAre(map[string]string{"WWW-Authenticate": "Bearer"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestPutRequestWithoutAuthorizationInfos_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("PUT", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseSpy := responseSpy{httptest.NewRecorder()}
	handlerSpy := &handlerSpy{}
	idp.HandleAuth(nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if err := responseSpy.assertHeadersAre(map[string]string{"WWW-Authenticate": "Bearer"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestDeleteRequestWithoutAuthorizationInfos_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseSpy := responseSpy{httptest.NewRecorder()}
	handlerSpy := &handlerSpy{}
	idp.HandleAuth(nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if err := responseSpy.assertHeadersAre(map[string]string{"WWW-Authenticate": "Bearer"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestPatchRequestWithoutAuthorizationInfos_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("PATCH", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseSpy := responseSpy{httptest.NewRecorder()}
	handlerSpy := &handlerSpy{}
	idp.HandleAuth(nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if err := responseSpy.assertHeadersAre(map[string]string{"WWW-Authenticate": "Bearer"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func returnFromCtx(value string) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) { return value, nil }
}

func TestRequestWithBearerAuthorization_PopulatesContextWithPrincipalAndAuthsession(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "aXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e1"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestRequestWithLowerCaseBearerAuthorization_PopulatesContextWithPrincipalAndAuthsession(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "bXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e2"}
	req.Header.Set("Authorization", "bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestRequestWithAuthSessionIdCookie_PopulatesContextWithPrincipalAndAuthsession(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const base64EncodedAuthSessionId = "cXGxJeb0q%2b%2ffS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA%2fx7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf%2fi5eWuu7Duh%2bOTtTjMOt9w%3d%26Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	const authSessionId = "cXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e3"}
	req.AddCookie(&http.Cookie{Name: "AuthSessionId", Value: base64EncodedAuthSessionId})
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestRequestWithBadUrlEncodedAuthSessionIdCookie_ReturnsStatus500(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const base64EncodedAuthSessionId = "abc%XX"
	const authSessionId = "id"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e3"}
	req.AddCookie(&http.Cookie{Name: "AuthSessionId", Value: base64EncodedAuthSessionId})
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestRequestWithBearerTokenAndCookie_PopulatesContextUsingBearerToken(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "dXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e3"}
	const cookieValue = "abcd"
	req.AddCookie(&http.Cookie{Name: "AuthSessionId", Value: cookieValue})
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestRequestWithBadTokenAndExternalValidationIsNotAllowed_RedirectsToIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	// token nicht bekannt oder abgelaufen
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	handlerSpy := handlerSpy{}
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// wenn token ungültig (nicht bekannt oder abgelaufen) dann schickt IdP ein 401
		http.Error(w, "", http.StatusUnauthorized)
	}))
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusFound); err != nil {
		t.Error(err)
	}
	if err := spy.assertHeadersAre(map[string]string{"Location": "/identityprovider/login?redirect=%2Fmyresource%2Fsubresource%3Fquery1%3Dabc%26query2%3D123"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestRequestWithBadTokenAndExternalValidationIsAllowed_RedirectsToIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	// token nicht bekannt oder abgelaufen
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	handlerSpy := handlerSpy{}
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// wenn token ungültig (nicht bekannt oder abgelaufen) dann schickt IdP ein 401
		http.Error(w, "", http.StatusUnauthorized)
	}))
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), true, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusFound); err != nil {
		t.Error(err)
	}
	if err := spy.assertHeadersAre(map[string]string{"Location": "/identityprovider/login?redirect=%2Fmyresource%2Fsubresource%3Fquery1%3Dabc%26query2%3D123"}); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestGetSystemBaseUriFromCtxReturnsError_ReturnsStatus500(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "eXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e4"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(func(ctx context.Context) (string, error) { return "", errors.New("any error") }, returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestGetTenantIdFromCtxReturnsError_ReturnsStatus500(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "fXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e4"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), func(ctx context.Context) (string, error) { return "", errors.New("any error") }, false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestIdPReturnsStatus500_ReturnsStatus500(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "gXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=1800, private")
		http.Error(w, "", http.StatusInternalServerError)
	}))
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestIdPReturnsMalformedJson_ReturnsStatus500(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "hXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=1800, private")
		w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
		_, _ = fmt.Fprint(w, `{"wrong":"json}`)
	}))
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestIdPReturnsPrincipalWithEmptyId_ReturnsStatus500(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "iXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=1800, private")
		w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(scim.Principal{})
	}))
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestUserIsCachedAndCacheEntryIsNotExpired_ReturnsCachedEntry(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "jXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e5"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set("Cache-Control", "max-age=1800, private")
		w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(1 * time.Nanosecond)
	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if idpCalled != 1 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 1)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestUserIsCachedButCacheEntryIsExpired_CallsIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "kXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e6"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set("Cache-Control", "max-age=1, private")
		w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(1 * time.Second)
	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if idpCalled != 2 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 2)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestUserIsCachedForDifferentTenant_CallsIdp(t *testing.T) {
	const authSessionId = "lXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

	reqTenant1, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	principalT1 := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e7"}
	reqTenant1.Header.Set("Authorization", "Bearer "+authSessionId)
	idpStub1 := newIdpStub(map[string]scim.Principal{authSessionId: principalT1}, nil)
	defer idpStub1.Close()
	handlerSpy1 := handlerSpy{}
	idp.HandleAuth(returnFromCtx(idpStub1.URL), returnFromCtx("1"), false, log, log)(&handlerSpy1).ServeHTTP(httptest.NewRecorder(), reqTenant1)

	if err := handlerSpy1.assertPrincipalIs(principalT1); err != nil {
		t.Error(err)
	}

	reqTenant2, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	principalT2 := scim.Principal{Id: "0bbbf1b6-017a-449a-ad5f-9723d28223e7"}
	reqTenant2.Header.Set("Authorization", "Bearer "+authSessionId)
	idpStub2 := newIdpStub(map[string]scim.Principal{authSessionId: principalT2}, nil)
	defer idpStub2.Close()
	handlerSpy2 := handlerSpy{}
	idp.HandleAuth(returnFromCtx(idpStub2.URL), returnFromCtx("2"), false, log, log)(&handlerSpy2).ServeHTTP(httptest.NewRecorder(), reqTenant2)

	if err := handlerSpy2.assertPrincipalIs(principalT2); err != nil {
		t.Error(err)
	}
}

func TestRequestAsExternalUserAndExternalUserValidationIsNotAllowed_ReturnsStatus403(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "hXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := newIdpStub(nil, map[string]scim.Principal{authSessionId: {Emails: []scim.UserValue{{"info@d-velop.de"}}, Groups: []scim.UserGroup{{Value: "3E093BE5-CCCE-435D-99F8-544656B98681"}}}})
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)
	if err := spy.assertStatusCodeIs(http.StatusForbidden); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestRequestAsExternalUserAndExternalValidationIsAllowed_PopulatesContextWithPrincipalAndAuthsession(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "1XGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Emails: []scim.UserValue{{"info@d-velop.de"}}, Groups: []scim.UserGroup{{Value: "3E093BE5-CCCE-435D-99F8-544656B98681"}}}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := new(handlerSpy)
	idpStub := newIdpStub(nil, map[string]scim.Principal{authSessionId: principal})
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), true, log, log)(handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestRequestAsInternalUserAndExternalValidationIsAllowed_PopulatesContextWithAppPrincipalAndAuthsession(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "2XGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "7bbbf1b6-017a-449a-ad5f-9723d28223e1"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := new(handlerSpy)
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), true, log, log)(handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestRequestAsApp_PopulatesContextWithPrincipalAndAuthsession(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "2XGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWasdfJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "some-app@app.idp.d-velop.local"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := new(handlerSpy)
	idpStub := newIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), true, log, log)(handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}

	if err := handlerSpy.assertPrincipalIsApp("some-app"); err != nil {
		t.Error(err)
	}
}

func TestIdpSendsNoCacheHeader_CallsIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "mXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e7"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)
	idp.HandleAuth(returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if idpCalled != 2 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 2)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestGetRequestWithoutAuthorizationInfosWithAcceptHeader(t *testing.T) {

	var testCases = []struct {
		name             string
		header           string
		expectedRedirect bool
	}{
		{"empty", "", true},
		{"exists", "text/", false},
		{"wildcard,text", "text/*", true},
		{"wildcard,any", "*/*", true},
		{"wildcard,secondary", "application/json; q=1.0, */*; q=0.8", true},
		{"exists", "text/html", true},
		{"unknown-mimetype", "something/else", false},
		{"exists,q=1", "text/html; q=1", true},
		{"exists,q=1.0", "text/html; q=1.0", true},
		{"exists,q=0.9", "text/html; q=0.9", true},
		{"exists,q=0", "text/html; q=0", false},
		{"exists,q=0.0", "text/html; q=0.0", false},
		{"missing", "application/json", false},
		{"secondary,q=0.9", "application/json; q=1.0, text/html; q=0.9", true},
		{"secondary,q=0", "application/json; q=1.0, text/html; q=0", false},
		{"secondary,q=1.0", "application/json; q=0.9, text/html; q=1.0", true},
		{"secondary,q=0.0", "application/json; q=1.0, text/html; q=0.", false}, // broken header
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
			if err != nil {
				t.Fatal(err)
			}
			if testCase.header != "" {
				req.Header.Add("accept", testCase.header)
			}

			responseSpy := responseSpy{httptest.NewRecorder()}

			handlerSpy := &handlerSpy{}
			idp.HandleAuth(nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

			if handlerSpy.hasBeenCalled {
				t.Error("inner handler should not have been called")
			}

			if testCase.expectedRedirect {
				if err := responseSpy.assertStatusCodeIs(http.StatusFound); err != nil {
					t.Error(err)
				}

				if err := responseSpy.assertHeadersAre(map[string]string{"Location": "/identityprovider/login?redirect=%2Fmyresource%2Fsubresource%3Fquery1%3Dabc%26query2%3D123"}); err != nil {
					t.Error(err)
				}
			} else {

				if err := responseSpy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
					t.Error(err)
				}
			}

		})
	}
}

func newIdpStub(principals map[string]scim.Principal, externalPrincipals map[string]scim.Principal) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/identityprovider/validate" {
			var bearerTokenRegex = regexp.MustCompile("^(?i)bearer (.*)$")
			authorizationHeader := r.Header.Get("Authorization")
			authToken := bearerTokenRegex.FindStringSubmatch(authorizationHeader)[1]

			w.Header().Set("Cache-Control", "max-age=1800, private")
			w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")

			if r.URL.RawQuery == "allowExternalValidation=true" {
				if externalPrincipal, exist := externalPrincipals[authToken]; exist {
					_ = json.NewEncoder(w).Encode(externalPrincipal)
				} else if principal, exist := principals[authToken]; exist {
					_ = json.NewEncoder(w).Encode(principal)
				} else {
					http.Error(w, "", http.StatusUnauthorized)
				}
			} else {
				if principal, exist := principals[authToken]; exist {
					_ = json.NewEncoder(w).Encode(principal)
				} else if _, exist := externalPrincipals[authToken]; exist {
					http.Error(w, "token is for external user", http.StatusForbidden)
				} else {
					http.Error(w, "", http.StatusUnauthorized)
				}
			}
			return
		}
		http.Error(w, "", http.StatusNotFound)
	}))
}

type handlerSpy struct {
	authSessionId string
	prinicpal     scim.Principal
	hasBeenCalled bool
	app           string
	isApp         bool
}

func (spy *handlerSpy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	spy.hasBeenCalled = true
	spy.authSessionId, _ = idp.AuthSessionIdFromCtx(r.Context())
	spy.prinicpal, _ = idp.PrincipalFromCtx(r.Context())
	spy.app, spy.isApp = idp.AppFromCtx(r.Context())
}

func (spy *handlerSpy) assertAuthSessionIdIs(expectedAuthSessionID string) error {
	if spy.authSessionId != expectedAuthSessionID {
		return fmt.Errorf("handler set wrong AuthSessionId on context: got %v want %v", spy.authSessionId, expectedAuthSessionID)
	}
	return nil
}

func (spy *handlerSpy) assertPrincipalIs(expectedPrincipal scim.Principal) error {
	if !reflect.DeepEqual(spy.prinicpal, expectedPrincipal) {
		return fmt.Errorf("handler set wrong principal on context: got \n %v want\n %v", spy.prinicpal, expectedPrincipal)
	}
	return nil
}

func (spy *handlerSpy) assertPrincipalIsApp(expectedApp string) error {
	if !spy.isApp {
		return fmt.Errorf("handler set non-app principal, got %+v", spy.prinicpal)
	}
	if spy.app != expectedApp {
		return fmt.Errorf("handler set wrong app, got %v, want %v", spy.app, expectedApp)
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

func (spy *responseSpy) assertHeadersAre(expectedHeaders map[string]string) error {
	for header, expectedValue := range expectedHeaders {
		if actualValue := spy.Header().Get(header); actualValue != expectedValue {
			return fmt.Errorf("handler returned wrong %v header: got %v want %v", header, actualValue, expectedValue)
		}
	}
	return nil
}

func log(ctx context.Context, logmessage string) {
	_ = ctx
	fmt.Println(logmessage)
}
