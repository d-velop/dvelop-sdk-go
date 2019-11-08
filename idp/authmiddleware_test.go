package idp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/d-velop/dvelop-sdk-go/idp"
	"github.com/d-velop/dvelop-sdk-go/idp/idpclient"
	"github.com/d-velop/dvelop-sdk-go/idp/test"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

var idpClient, _ = idpclient.New()

const validAuthSessionId = "aXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

var principals = map[string]scim.Principal{
	validAuthSessionId: {Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e1"},
}

const validExternalAuthSessionId = "1XGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Cnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

var externalPrincipals = map[string]scim.Principal{
	validExternalAuthSessionId: {Emails: []scim.UserValue{{"info@d-velop.de"}}, Groups: []scim.UserGroup{{Value: "3E093BE5-CCCE-435D-99F8-544656B98681"}}},
}

func TestNoAuthSessionId(t *testing.T) {
	type result struct {
		StatusCode int
		Headers    http.Header
	}
	testcases := map[string]struct {
		method  string
		headers map[string]string
		url     string
		cookie  *http.Cookie
		want    result
	}{
		// read function name and testcase name as one sentence. e.g. TestNoAuthSessionIdAndHeadRequestAndHtmlAccepted_Middleware_RedirectsToIdp
		"AndHeadRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodHead, headers: map[string]string{"Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{"Location": {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"AndGetRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodGet, headers: map[string]string{"Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{"Location": {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"ButBasicAuthorizationAndGetRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodGet, headers: map[string]string{"Authorization": "Basic adadbk", "Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{"Location": {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"ButOtherCookieAndGetRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodGet, cookie: &http.Cookie{Name: "AnyCookie", Value: "adadbk"}, headers: map[string]string{"Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{"Location": {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"AndPostRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPost, headers: map[string]string{"Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndPutRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPut, headers: map[string]string{"Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndDeleteRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodDelete, headers: map[string]string{"Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndPatchRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPatch, headers: map[string]string{"Accept": "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			for key, val := range tc.headers {
				req.Header.Add(key, val)
			}
			responseSpy := responseSpy{httptest.NewRecorder()}
			handlerSpy := &handlerSpy{}

			idp.Authenticate(idpClient, nil, nil, false, discardLog, discardLog)(handlerSpy).ServeHTTP(responseSpy, req)

			got := result{
				StatusCode: responseSpy.Code,
				Headers:    responseSpy.Header(),
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("\nexpected: %v\ngot     : %v", tc.want, got)
			}
			if handlerSpy.hasBeenCalled {
				t.Error("inner handler should not have been called")
			}
		})
	}
}

func TestValidAuthSessionId(t *testing.T) {
	type result struct {
		Principal     scim.Principal
		AuthSessionId string
	}
	testcases := map[string]struct {
		method  string
		headers map[string]string
		url     string
		cookie  *http.Cookie
		want    result
	}{
		// read function name and testcase name as one sentence. e.g. TestValidAuthSessionIdAsBearerTokenAndGetRequest_Middleware_PopulatesContextWithPrincipalAndAuthSession
		"AsBearerTokenAndGetRequest_Middleware_PopulatesContextWithPrincipalAndAuthSession": {
			method: http.MethodGet, headers: map[string]string{"Authorization": "Bearer " + validAuthSessionId}, url: "/a/b?q1=x&q2=1",
			want: result{Principal: principals[validAuthSessionId], AuthSessionId: validAuthSessionId}},
		"AsLowerCaseBearerTokenAndGetRequest_Middleware_PopulatesContextWithPrincipalAndAuthSession": {
			method: http.MethodGet, headers: map[string]string{"Authorization": "bearer " + validAuthSessionId}, url: "/a/b?q1=x&q2=1",
			want: result{Principal: principals[validAuthSessionId], AuthSessionId: validAuthSessionId}},
		"AsCookieAndGetRequest_Middleware_PopulatesContextWithPrincipalAndAuthSession": {
			method: http.MethodGet, cookie: &http.Cookie{Name: "AuthSessionId", Value: url.QueryEscape(validAuthSessionId)}, url: "/a/b?q1=x&q2=1",
			want: result{Principal: principals[validAuthSessionId], AuthSessionId: validAuthSessionId}},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			for key, val := range tc.headers {
				req.Header.Add(key, val)
			}
			handlerSpy := &handlerSpy{}
			idpStub := test.NewIdpStub(principals, externalPrincipals)
			defer idpStub.Close()

			idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, discardLog, discardLog)(handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

			got := result{
				Principal:     handlerSpy.principal,
				AuthSessionId: handlerSpy.authSessionId,
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("\nexpected: %v\ngot     : %v", tc.want, got)
			}
		})
	}
}

func returnFromCtx(value string) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) { return value, nil }
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
	idpStub := test.NewIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

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
	idpStub := test.NewIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

type validatorStub struct {
	returnedPrincipal *scim.Principal
}

func (v *validatorStub) Validate(ctx context.Context, systemBaseUri string, tenantId string, authSessionId string) (*scim.Principal, error) {
	return v.returnedPrincipal, nil
}

func TestGetRequestWithInvalidTokenAndHtmlAccepted_RedirectsToIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	req.Header.Set("Accept", "text/html")
	handlerSpy := handlerSpy{}
	idpClientStub := &validatorStub{returnedPrincipal: nil}
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClientStub, returnFromCtx(""), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

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

func TestGetRequestWithInvalidTokenAndNoHtmlAccepted_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	req.Header.Set("Accept", "application/json")
	handlerSpy := handlerSpy{}
	idpClientStub := &validatorStub{returnedPrincipal: nil}
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClientStub, returnFromCtx(""), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestPostRequestWithInvalidToken_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("POST", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	handlerSpy := handlerSpy{}
	idpClientStub := &validatorStub{returnedPrincipal: nil}
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClientStub, returnFromCtx(""), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestPutRequestWithInvalidToken_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("PUT", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	handlerSpy := handlerSpy{}
	idpClientStub := &validatorStub{returnedPrincipal: nil}
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClientStub, returnFromCtx(""), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestDeleteRequestWithInvalidToken_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	handlerSpy := handlerSpy{}
	idpClientStub := &validatorStub{returnedPrincipal: nil}
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClientStub, returnFromCtx(""), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestPatchRequestWithInvalidToken_ReturnsStatus401(t *testing.T) {
	req, err := http.NewRequest("PATCH", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	handlerSpy := handlerSpy{}
	idpClientStub := &validatorStub{returnedPrincipal: nil}
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClientStub, returnFromCtx(""), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

	if err := spy.assertStatusCodeIs(http.StatusUnauthorized); err != nil {
		t.Error(err)
	}
	if handlerSpy.hasBeenCalled {
		t.Error("inner handler should not have been called")
	}
}

func TestGetRequestWithInvalidTokenAndAndHtmlAcceptedAndExternalValidation_RedirectsToIdp(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const badToken = "200e7388-1834-434b-be79-3745181e1457"
	req.Header.Set("Authorization", "Bearer "+badToken)
	req.Header.Set("Accept", "text/html")
	handlerSpy := handlerSpy{}
	idpClientStub := &validatorStub{returnedPrincipal: nil}
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClientStub, returnFromCtx(""), returnFromCtx("1"), true, log, log)(&handlerSpy).ServeHTTP(spy, req)

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
	idpStub := test.NewIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClient, func(ctx context.Context) (string, error) { return "", errors.New("any error") }, returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

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
	idpStub := test.NewIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), func(ctx context.Context) (string, error) { return "", errors.New("any error") }, false, log, log)(&handlerSpy).ServeHTTP(spy, req)

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
		http.Error(w, "", http.StatusInternalServerError)
	}))
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

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

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)

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

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(1 * time.Nanosecond)
	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

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

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(1 * time.Second)
	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

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
	idpStub1 := test.NewIdpStub(map[string]scim.Principal{authSessionId: principalT1}, nil)
	defer idpStub1.Close()
	handlerSpy1 := handlerSpy{}
	idp.Authenticate(idpClient, returnFromCtx(idpStub1.URL), returnFromCtx("1"), false, log, log)(&handlerSpy1).ServeHTTP(httptest.NewRecorder(), reqTenant1)

	if err := handlerSpy1.assertPrincipalIs(principalT1); err != nil {
		t.Error(err)
	}

	reqTenant2, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	principalT2 := scim.Principal{Id: "0bbbf1b6-017a-449a-ad5f-9723d28223e7"}
	reqTenant2.Header.Set("Authorization", "Bearer "+authSessionId)
	idpStub2 := test.NewIdpStub(map[string]scim.Principal{authSessionId: principalT2}, nil)
	defer idpStub2.Close()
	handlerSpy2 := handlerSpy{}
	idp.Authenticate(idpClient, returnFromCtx(idpStub2.URL), returnFromCtx("2"), false, log, log)(&handlerSpy2).ServeHTTP(httptest.NewRecorder(), reqTenant2)

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
	idpStub := test.NewIdpStub(nil, map[string]scim.Principal{authSessionId: {Emails: []scim.UserValue{{"info@d-velop.de"}}, Groups: []scim.UserGroup{{Value: "3E093BE5-CCCE-435D-99F8-544656B98681"}}}})
	defer idpStub.Close()
	spy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(spy, req)
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
	idpStub := test.NewIdpStub(nil, map[string]scim.Principal{authSessionId: principal})
	defer idpStub.Close()

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), true, log, log)(handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
		t.Error(err)
	}
}

func TestRequestAsInternalUserAndExternalValidationIsAllowed_PopulatesContextWithPrincipalAndAuthsession(t *testing.T) {
	req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	const authSessionId = "2XGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	principal := scim.Principal{Id: "7bbbf1b6-017a-449a-ad5f-9723d28223e1"}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	handlerSpy := new(handlerSpy)
	idpStub := test.NewIdpStub(map[string]scim.Principal{authSessionId: principal}, nil)
	defer idpStub.Close()

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), true, log, log)(handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if err := handlerSpy.assertAuthSessionIdIs(authSessionId); err != nil {
		t.Error(err)
	}
	if err := handlerSpy.assertPrincipalIs(principal); err != nil {
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

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)
	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

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
			idp.Authenticate(idpClient, nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

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

type handlerSpy struct {
	authSessionId string
	principal     scim.Principal
	hasBeenCalled bool
}

func (spy *handlerSpy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	spy.hasBeenCalled = true
	spy.authSessionId, _ = idp.AuthSessionIdFromCtx(r.Context())
	spy.principal, _ = idp.PrincipalFromCtx(r.Context())
}

func (spy *handlerSpy) assertAuthSessionIdIs(expectedAuthSessionID string) error {
	if spy.authSessionId != expectedAuthSessionID {
		return fmt.Errorf("handler set wrong AuthSessionId on context: got %v want %v", spy.authSessionId, expectedAuthSessionID)
	}
	return nil
}

func (spy *handlerSpy) assertPrincipalIs(expectedPrincipal scim.Principal) error {
	if !reflect.DeepEqual(spy.principal, expectedPrincipal) {
		return fmt.Errorf("handler set wrong principal on context: got \n %v want\n %v", spy.principal, expectedPrincipal)
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

func discardLog(ctx context.Context, logmessage string) {
	_ = ctx
	_ = logmessage
}
