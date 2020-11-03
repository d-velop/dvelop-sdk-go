package idpclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/google/go-cmp/cmp"

	"github.com/d-velop/dvelop-sdk-go/idp/idpclient"
	"github.com/d-velop/dvelop-sdk-go/idp/scim"
	"github.com/d-velop/dvelop-sdk-go/idp/test"
)

func ExampleNew_default() {
	// create client with defaults
	_, _ = idpclient.New()
}

func ExampleNew_custom() {
	// create client with custom http.Client
	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}
	_, _ = idpclient.New(idpclient.HttpClient(httpClient))
}

var defaultClient, _ = idpclient.New()

const validAuthSessionId = "aXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

var principals = map[string]scim.Principal{
	validAuthSessionId: {Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e1"},
}

const validExternalAuthSessionId = "1XGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Cnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

var externalPrincipals = map[string]scim.Principal{
	validExternalAuthSessionId: {Emails: []scim.UserValue{{"info@d-velop.de"}}, Groups: []scim.UserGroup{{Value: "3E093BE5-CCCE-435D-99F8-544656B98681"}}},
}

const invalidAuthSessionId = "2XGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Dnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

func TestValidAuthSessionIdOfInternalUser_Validate_ReturnsInternalPrincipal(t *testing.T) {
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
	defer idpStub.Close()

	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", validAuthSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, principals[validAuthSessionId]) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, principals[validAuthSessionId])
	}
	if p.IsExternal() {
		t.Errorf("validate returned external principal \n %v \n but principal should not be external.", p)
	}
}

func TestValidAuthSessionIdOfExternalUser_Validate_ReturnsExternalPrincipal(t *testing.T) {
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
	defer idpStub.Close()

	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", validExternalAuthSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, externalPrincipals[validExternalAuthSessionId]) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, externalPrincipals[validExternalAuthSessionId])
	}
	if p.IsExternal() == false {
		t.Errorf("validate returned no internal principal \n %v \n but expected external principal.", p)
	}
}

func TestAuthSessionIdIsInvalid_Validate_ReturnsNilPrincipal(t *testing.T) {
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
	defer idpStub.Close()

	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", invalidAuthSessionId)

	if err != nil {
		t.Error(err)
	}
	if p != nil {
		t.Errorf("Expected validate to return nil principal but got:\n %v", p)
	}
}

func TestIdpReturnsStatus500_Validate_ReturnsErrorAndNilPrincipal(t *testing.T) {
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "a fatal error occurred", http.StatusInternalServerError)
	}))
	defer idpStub.Close()

	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", invalidAuthSessionId)

	if p != nil {
		t.Errorf("Expected validate to return nil principal but got:\n %v", p)
	}
	if err == nil {
		t.Error("Expected validate to return an error but error was nil")
	}
}

func TestIdpReturnsMalformedJson_Validate_ReturnsErrorAndNilPrincipal(t *testing.T) {
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headers.CacheControl, "max-age=1800, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
		_, _ = fmt.Fprint(w, `{"wrong":"json}`)
	}))
	defer idpStub.Close()

	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", invalidAuthSessionId)

	if p != nil {
		t.Errorf("Expected validate to return nil principal but got:\n %v", p)
	}
	if err == nil {
		t.Error("Expected validate to return an error but error was nil")
	}
}

func TestPrincipalIsCachedAndCacheEntryIsNotExpired_Validate_ReturnsCachedEntry(t *testing.T) {
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e5"}
	const authSessionId = "11GxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.CacheControl, "max-age=1800, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	if _, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId); err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Nanosecond)
	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, principal) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, principal)
	}
	if idpCalled != 1 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 1)
	}
}

func TestPrincipalIsCachedButCacheEntryIsExpired_Validate_CallsIdp(t *testing.T) {
	principal := scim.Principal{Id: "fffff1b6-017a-449a-ad5f-9723d28223e5"}
	const authSessionId = "22GxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.CacheControl, "max-age=1, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	if _, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId); err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)
	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, principal) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, principal)
	}
	if idpCalled != 2 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 2)
	}
}

func TestPrincipalIsCachedForDifferentTenant_Validate_CallsIdp(t *testing.T) {
	principal := scim.Principal{Id: "9bbbf1b6-017a-449a-ad5f-9723d28223e5"}
	const authSessionId = "33GxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.CacheControl, "max-age=1800, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	if _, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId); err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Nanosecond)
	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "2", authSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, principal) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, principal)
	}
	if idpCalled != 2 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 2)
	}
}

func TestIdpSentsNoCacheHeader_Validate_CallsIdp(t *testing.T) {
	principal := scim.Principal{Id: "fffff1b6-017a-449a-ad5f-9723d28223e5"}
	const authSessionId = "44GxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	if _, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId); err != nil {
		t.Error(err)
	}
	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, principal) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, principal)
	}
	if idpCalled != 2 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 2)
	}
}

func TestIdpSentsMaxAgeZero_Validate_CallsIdp(t *testing.T) {
	principal := scim.Principal{Id: "fffff1b6-017a-449a-ad5f-9723d28223e5"}
	const authSessionId = "55GxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	idpCalled := 0
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpCalled++
		w.Header().Set(headers.CacheControl, "max-age=0, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()

	if _, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId); err != nil {
		t.Error(err)
	}
	p, err := defaultClient.Validate(context.Background(), idpStub.URL, "1", authSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, principal) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, principal)
	}
	if idpCalled != 2 {
		t.Errorf("IdP has been called %v times but expected %v times", idpCalled, 2)
	}
}

func TestContextWithTimeoutAndRequestTimedOut_Validate_ReturnsTimeout(t *testing.T) {
	principal := scim.Principal{Id: "fffff1b6-017a-449a-ad5f-9723d28223e5"}
	const authSessionId = "88GxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.Header().Set(headers.CacheControl, "max-age=1600, private")
		w.Header().Set(headers.ContentType, "application/hal+json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(principal)
		return
	}))
	defer idpStub.Close()
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var urlError *url.Error
	if _, err := defaultClient.Validate(ctxWithTimeout, idpStub.URL, "1", authSessionId); err == nil {
		t.Error("Expected validate to return an *url.Error because of a timeout but validate didn't return an error")
	} else if !(errors.As(err, &urlError) && urlError.Timeout()) {
		t.Errorf("Expected validate to to return an *url.Error because of a timeout but validate returned another error %v", err)
	}
}

func TestContextWithTimeoutAndRequestDoesntTimeOut_Validate_ReturnsPrincipal(t *testing.T) {
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
	defer idpStub.Close()
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p, err := defaultClient.Validate(ctxWithTimeout, idpStub.URL, "1", validAuthSessionId)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*p, principals[validAuthSessionId]) {
		t.Errorf("validate returned wrong principal: got \n %v want\n %v", p, principals[validAuthSessionId])
	}
	if p.IsExternal() {
		t.Errorf("validate returned external principal \n %v \n but principal should not be external.", p)
	}
}

type RoundTripperSpy struct {
	HasBeenCalled bool
}

func (h *RoundTripperSpy) RoundTrip(req *http.Request) (*http.Response, error) {
	h.HasBeenCalled = true
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
		Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
	}, nil
}

func TestCustomHttpClientSpecified_New_UsesCustomHttpClient(t *testing.T) {
	spy := RoundTripperSpy{}
	client, err := idpclient.New(idpclient.HttpClient(&http.Client{
		Transport: &spy,
	}))
	if err != nil {
		t.Error(err)
	}

	_, _ = client.Validate(context.Background(), "URL.Doesnt.Matter.Because.No.Http.Call.Is.Made", "1", validAuthSessionId)

	if !spy.HasBeenCalled {
		t.Error("Expected custom http client to be used. But custom http client wasn't used")
	}
}

type PrincipalCacheSpy struct {
	Invocations int
}

func (pc *PrincipalCacheSpy) Get(key string) (item interface{}, found bool) {
	pc.Invocations++
	return nil, false
}

func (pc *PrincipalCacheSpy) Set(key string, item interface{}, cacheDuration time.Duration) {
	pc.Invocations++
}

func TestCustomPrincipalCacheSpecified_New_UsesCustomPrincipalCache(t *testing.T) {
	idpStub := test.NewIdpValidateStub(principals, externalPrincipals)
	defer idpStub.Close()

	spy := &PrincipalCacheSpy{}
	client, e := idpclient.New(idpclient.PrincipalCache(spy))
	if e != nil {
		t.Error(e)
	}

	_, err := client.Validate(context.Background(), idpStub.URL, "1", validAuthSessionId)

	if err != nil {
		t.Error(err)
	}
	if spy.Invocations <= 0 {
		t.Error("Expected custom cache to be used. But custom cache wasn't used")
	}
}

func TestCallerIsAuthorizedAndPrincipalExists_GetPrincipalById_ReturnsPrincipal(t *testing.T) {
	const authSessionIdFromAuthorizedCaller = validAuthSessionId
	existingPrincipal := scim.Principal{Id: "719052ec-0c46-4db4-9cc4-f57e6492d25d"}
	idpStub := test.NewIdpUsersStub(authSessionIdFromAuthorizedCaller, existingPrincipal)

	got, err := defaultClient.GetPrincipalById(context.Background(), idpStub.URL, "1", authSessionIdFromAuthorizedCaller, existingPrincipal.Id)

	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(&existingPrincipal, got); diff != "" {
		t.Errorf("\nexpected: %v\ngot     : %v", &existingPrincipal, got)
	}
}

func TestCallerIsAuthorizedAndPrincipalDoesntExist_GetPrincipalById_ReturnsError(t *testing.T) {
	const authSessionIdFromAuthorizedCaller = validAuthSessionId
	existingPrincipal := scim.Principal{Id: "719052ec-0c46-4db4-9cc4-f57e6492d25d"}
	idpStub := test.NewIdpUsersStub(authSessionIdFromAuthorizedCaller, existingPrincipal)

	noneExistingPrincipalId := "83db85b2-89d3-4586-b455-ad041ff38195"
	got, err := defaultClient.GetPrincipalById(context.Background(), idpStub.URL, "1", authSessionIdFromAuthorizedCaller, noneExistingPrincipalId)

	if err == nil || got != nil {
		t.Errorf("expected an error because principal with id '%s' doesn't exist but got no error", noneExistingPrincipalId)
	}
}

func TestCallerNotAuthorizedAndPrincipalExists_GetPrincipalById_ReturnsError(t *testing.T) {
	const authSessionIdFromUnauthorizedCaller = invalidAuthSessionId
	const authSessionIdFromAuthorizedCaller = validAuthSessionId
	existingPrincipal := scim.Principal{Id: "719052ec-0c46-4db4-9cc4-f57e6492d25d"}
	idpStub := test.NewIdpUsersStub(authSessionIdFromAuthorizedCaller, existingPrincipal)

	got, err := defaultClient.GetPrincipalById(context.Background(), idpStub.URL, "1", authSessionIdFromUnauthorizedCaller, existingPrincipal.Id)

	if err == nil || got != nil {
		t.Error("expected an error because caller is not authorized to call Idp but got no error")
	}
}

func TestIdpReturnsUnexpectedStatusCode_GetPrincipalById_ReturnsError(t *testing.T) {
	const authSessionIdFromAuthorizedCaller = validAuthSessionId
	existingPrincipal := scim.Principal{Id: "719052ec-0c46-4db4-9cc4-f57e6492d25d"}
	idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusConflict)
	}))
	got, err := defaultClient.GetPrincipalById(context.Background(), idpStub.URL, "1", authSessionIdFromAuthorizedCaller, existingPrincipal.Id)

	if err == nil || got != nil {
		t.Error("expected an error because idp returned unexpected http error")
	}
}

func Test_GetPrincipalById_Errors(t *testing.T) {
	authSessionId := "authSessionId"
	tenantId := "tenantId"
	principalId := "principalId"

	tests := []struct {
		name       string
		statusCode int
	}{
		{"IdpForbiddenError", http.StatusForbidden},
		{"IdpNotFoundError", http.StatusNotFound},
		{"IdpInternalServerError", http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "error", test.statusCode)
			}))

			got, err := defaultClient.GetPrincipalById(context.Background(), idpStub.URL, tenantId, authSessionId, principalId)

			if got != nil {
				t.Error("expected nil principal because idp returned unexpected http error")
			}

			switch err.(type) {
			case idpclient.IdpClientError:
			default:
				t.Errorf("unexpected error - expected %T, but is %T", idpclient.IdpClientError{}, err)
			}

			if err.(idpclient.IdpClientError).StatusCode != test.statusCode {
				t.Errorf("unexpected status code - expected %d, got %d", err.(idpclient.IdpClientError).StatusCode, test.statusCode)
			}
		})
	}
}

func Test_Validate_Errors(t *testing.T) {
	authSessionId := "authSessionId"
	tenantId := "tenantId"

	tests := []struct {
		name       string
		statusCode int
	}{
		{"IdpUnauthorizedError", http.StatusUnauthorized},
		{"IdpNotFoundError", http.StatusNotFound},
		{"IdpInternalServerError", http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			idpStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "error", test.statusCode)
			}))

			got, err := defaultClient.Validate(context.Background(), idpStub.URL, tenantId, authSessionId)

			if got != nil {
				t.Error("expected nil principal because idp returned unexpected http error")
			}

			if test.statusCode == http.StatusUnauthorized {
				if err != nil {
					t.Errorf("did not expect error, but got %v", err)
				}
			} else {
				switch err.(type) {
				case idpclient.IdpClientError:
				default:
					t.Errorf("unexpected error - expected %T, but is %T", idpclient.IdpClientError{}, err)
				}

				if err.(idpclient.IdpClientError).StatusCode != test.statusCode {
					t.Errorf("unexpected status code - expected %d, got %d", err.(idpclient.IdpClientError).StatusCode, test.statusCode)
				}
			}
		})
	}
}
