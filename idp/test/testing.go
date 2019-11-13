package test

import (
	"encoding/json"
	"github.com/d-velop/dvelop-sdk-go/idp/scim"
	"net/http"
	"net/http/httptest"
	"regexp"
)

var bearerTokenRegex = regexp.MustCompile("^(?i)bearer (.*)$")

func NewIdpValidateStub(principals map[string]scim.Principal, externalPrincipals map[string]scim.Principal) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/identityprovider/validate" {
			authorizationHeader := r.Header.Get("Authorization")
			authToken := bearerTokenRegex.FindStringSubmatch(authorizationHeader)[1]

			w.Header().Set("Cache-Control", "max-age=1800, private")
			w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")

			if r.URL.RawQuery == "allowExternalValidation=true" {
				if principal, exist := principals[authToken]; exist {
					_ = json.NewEncoder(w).Encode(principal)
				} else if externalPrincipal, exist := externalPrincipals[authToken]; exist {
					_ = json.NewEncoder(w).Encode(externalPrincipal)
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

func NewIdpUsersStub(authSessionIdFromAuthorizedCaller string, existingPrincipal scim.Principal) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/identityprovider/scim/users/"+existingPrincipal.Id {
			authorizationHeader := r.Header.Get("Authorization")
			authToken := bearerTokenRegex.FindStringSubmatch(authorizationHeader)[1]

			if authToken != authSessionIdFromAuthorizedCaller {
				http.Error(w, `{"msg":"user unauthorized"}`, http.StatusForbidden)
			} else {
				_ = json.NewEncoder(w).Encode(existingPrincipal)
			}
			return
		}
		http.Error(w, "", http.StatusNotFound)
	}))
}
