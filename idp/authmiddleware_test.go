package idp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/google/go-cmp/cmp"

	"github.com/d-velop/dvelop-sdk-go/idp"
	"github.com/d-velop/dvelop-sdk-go/idp/idpclient"
	"github.com/d-velop/dvelop-sdk-go/idp/test"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

var idpClient, _ = idpclient.New()

const validAuthSessionId = "aXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
const secondValidAuthSessionId = "bYGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

var principals = map[string]scim.Principal{
	validAuthSessionId:       {Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e1"},
	secondValidAuthSessionId: {Id: "1234f1b6-017a-449a-ad5f-9723d2822fff"},
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
		// read function name and testCase name as one sentence. e.g. TestNoAuthSessionIdAndHeadRequestAndHtmlAccepted_Middleware_RedirectsToIdp
		"AndGetRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodGet, headers: map[string]string{headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{headers.Location: {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"AndHeadRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodHead, headers: map[string]string{headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{headers.Location: {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"ButBasicAuthorizationAndGetRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodGet, headers: map[string]string{headers.Authorization: "Basic adadbk", headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{headers.Location: {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"ButOtherCookieAndGetRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodGet, cookie: &http.Cookie{Name: "AnyCookie", Value: "adadbk"}, headers: map[string]string{headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{headers.Location: {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"AndPostRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPost, headers: map[string]string{headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndPutRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPut, headers: map[string]string{headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndDeleteRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodDelete, headers: map[string]string{headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndPatchRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPatch, headers: map[string]string{headers.Accept: "text/html"}, url: "/a/b?q1=x&q2=1",
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

			idp.Authenticate(idpClient, nil, nil, false, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

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

func TestInvalidAuthSessionId(t *testing.T) {
	const invalidToken = "200e7388-1834-434b-be79-3745181e1457"
	type result struct {
		StatusCode int
		Headers    http.Header
	}
	testcases := map[string]struct {
		method                  string
		headers                 map[string]string
		url                     string
		cookie                  *http.Cookie
		want                    result
		allowExternalValidation bool
	}{
		// read function name and testCase name as one sentence. e.g. TestNoAuthSessionIdAndHeadRequestAndHtmlAccepted_Middleware_RedirectsToIdp
		"AndGetRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodGet, headers: map[string]string{headers.Accept: "text/html", headers.Authorization: "Bearer " + invalidToken}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{headers.Location: {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"AndHeadRequestAndHtmlAccepted_Middleware_RedirectsToIdp": {
			method: http.MethodHead, headers: map[string]string{headers.Accept: "text/html", headers.Authorization: "Bearer " + invalidToken}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{headers.Location: {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
		"AndGetRequestAndHtmlNotAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodGet, headers: map[string]string{headers.Accept: "application/json", headers.Authorization: "Bearer " + invalidToken}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndPostRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPost, headers: map[string]string{headers.Accept: "text/html", headers.Authorization: "Bearer " + invalidToken}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndPutRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPut, headers: map[string]string{headers.Accept: "text/html", headers.Authorization: "Bearer " + invalidToken}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndDeleteRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodDelete, headers: map[string]string{headers.Accept: "text/html", headers.Authorization: "Bearer " + invalidToken}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndPatchRequestAndHtmlAccepted_Middleware_ReturnsStatus401AndWWW-AuthenticateBearerHeader": {
			method: http.MethodPatch, headers: map[string]string{headers.Accept: "text/html", headers.Authorization: "Bearer " + invalidToken}, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusUnauthorized, Headers: http.Header{"Www-Authenticate": {"Bearer"}}}},
		"AndGetRequestAndHtmlAcceptedAndExternalValidation_Middleware_RedirectsToIdp": {
			method: http.MethodGet, headers: map[string]string{headers.Accept: "text/html", headers.Authorization: "Bearer " + invalidToken}, allowExternalValidation: true, url: "/a/b?q1=x&q2=1",
			want: result{StatusCode: http.StatusFound, Headers: http.Header{headers.Location: {"/identityprovider/login?redirect=" + url.QueryEscape("/a/b?q1=x&q2=1")}}}},
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
			idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
			defer idpStub.Close()

			idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), tc.allowExternalValidation, log, log)(handlerSpy).ServeHTTP(responseSpy, req)

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
		// read function name and testCase name as one sentence. e.g. TestValidAuthSessionIdAsBearerTokenAndGetRequest_Middleware_PopulatesContextWithPrincipalAndAuthSession
		"AsBearerToken_Middleware_PopulatesContextWithPrincipalAndAuthSession": {
			method: http.MethodGet, headers: map[string]string{headers.Authorization: "Bearer " + validAuthSessionId}, url: "/a/b?q1=x&q2=1",
			want: result{Principal: principals[validAuthSessionId], AuthSessionId: validAuthSessionId}},
		"AsLowerCaseBearerToken_Middleware_PopulatesContextWithPrincipalAndAuthSession": {
			method: http.MethodGet, headers: map[string]string{headers.Authorization: "bearer " + validAuthSessionId}, url: "/a/b?q1=x&q2=1",
			want: result{Principal: principals[validAuthSessionId], AuthSessionId: validAuthSessionId}},
		"AsCookie_Middleware_PopulatesContextWithPrincipalAndAuthSession": {
			method: http.MethodGet, cookie: &http.Cookie{Name: "AuthSessionId", Value: url.QueryEscape(validAuthSessionId)}, url: "/a/b?q1=x&q2=1",
			want: result{Principal: principals[validAuthSessionId], AuthSessionId: validAuthSessionId}},
		"AsBearerTokenAndAsCookie_Middleware_PopulatesContextWithPrincipalAndAuthSessionFromBearerToken": {
			method: http.MethodGet, headers: map[string]string{headers.Authorization: "Bearer " + validAuthSessionId}, cookie: &http.Cookie{Name: "AuthSessionId", Value: url.QueryEscape(secondValidAuthSessionId)}, url: "/a/b?q1=x&q2=1",
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
			idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
			defer idpStub.Close()

			idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

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
	const badUrlEncodedAuthSessionId = "abc%XX"
	req.AddCookie(&http.Cookie{Name: "AuthSessionId", Value: badUrlEncodedAuthSessionId})
	handlerSpy := handlerSpy{}
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
	defer idpStub.Close()
	responseSpy := responseSpy{httptest.NewRecorder()}

	idp.Authenticate(idpClient, returnFromCtx(idpStub.URL), returnFromCtx("1"), false, log, log)(&handlerSpy).ServeHTTP(responseSpy, req)

	if err := responseSpy.assertStatusCodeIs(http.StatusInternalServerError); err != nil {
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
	req.Header.Set(headers.Authorization, "Bearer "+validAuthSessionId)
	handlerSpy := handlerSpy{}
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
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
	req.Header.Set(headers.Authorization, "Bearer "+validAuthSessionId)
	handlerSpy := handlerSpy{}
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headers.CacheControl, "max-age=1800, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.CacheControl, "max-age=1800, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.CacheControl, "max-age=1, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
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
	reqTenant1.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	idpStub1 := test.NewIdpValidateStub(map[string]scim.Principal{authSessionId: principalT1}, nil)
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
	reqTenant2.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	idpStub2 := test.NewIdpValidateStub(map[string]scim.Principal{authSessionId: principalT2}, nil)
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpStub := test.NewIdpValidateStub(nil, map[string]scim.Principal{authSessionId: {Emails: []scim.UserValue{{"info@d-velop.de"}}, Groups: []scim.UserGroup{{Value: "3E093BE5-CCCE-435D-99F8-544656B98681"}}}})
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	handlerSpy := new(handlerSpy)
	idpStub := test.NewIdpValidateStub(nil, map[string]scim.Principal{authSessionId: principal})
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	handlerSpy := new(handlerSpy)
	idpStub := test.NewIdpValidateStub(map[string]scim.Principal{authSessionId: principal}, nil)
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
	req.Header.Set(headers.Authorization, "Bearer "+authSessionId)
	handlerSpy := handlerSpy{}
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
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

func TestNoAuthSessionIdAndGetRequestAndAcceptHeaderIs(t *testing.T) {

	var testCases = []struct {
		header           string
		expectedRedirect bool
	}{
		{"", true},
		{"text/", false},
		{"text/*", true},
		{"*/*", true},
		{"application/json; q=1.0, */*; q=0.8", true},
		{"text/html", true},
		{"something/else", false},
		{"text/html; q=1", true},
		{"text/html; q=1.0", true},
		{"text/html; q=0.9", true},
		{"text/html; q=0", false},
		{"text/html; q=0.0", false},
		{"application/json", false},
		{"application/json; q=1.0, text/html; q=0.9", true},
		{"application/json; q=1.0, text/html; q=0", false},
		{"application/json; q=0.9, text/html; q=1.0", true},
		{"application/json; q=1.0, text/html; q=0.", false}, // broken header
	}

	for _, testCase := range testCases {
		name := testCase.header
		if name == "" {
			name = "Empty"
		}
		t.Run(name, func(t *testing.T) {

			req, err := http.NewRequest("GET", "/myresource/subresource?query1=abc&query2=123", nil)
			if err != nil {
				t.Fatal(err)
			}
			if testCase.header != "" {
				req.Header.Add(headers.Accept, testCase.header)
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

				if err := responseSpy.assertHeadersAre(map[string]string{headers.Location: "/identityprovider/login?redirect=%2Fmyresource%2Fsubresource%3Fquery1%3Dabc%26query2%3D123"}); err != nil {
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
	_ = logmessage
}
