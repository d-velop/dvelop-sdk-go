package idpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

func TestPrincipalById(t *testing.T) {

	var testCases = []struct {
		name              string
		requestedUserId   string
		expectedPrincipal *scim.Principal
		wantError         bool
	}{
		{"UnknowUserId", "unknown", nil, true},
		{"NormalUserId", "abc-123", &scim.Principal{Id: "abc-123"}, false},
		{"UserIdWithSlash", "abc/123", nil, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(tt *testing.T) {
			const authSessionId = "dXGxJeb0q+/fS8biFi8FE7TovJPPEPyzlDxT6bh5p5pHA/x7CEi1w9egVhEMz8IWhrtvJRFnkSqJnLr61cOKf/i5eWuu7Duh+OTtTjMOt9w=&Bnh4NNU90wH_OVlgbzbdZOEu1aSuPlbUctiCdYTonZ3Ap_Zd3bVL79I-dPdHf4OOgO8NKEdqyLsqc8RhAOreXgJqXuqsreeI"

			idpStub := newIdpStub(testCase.expectedPrincipal)

			returnedPrincipal, err := PrincipalById(authSessionId, idpStub.URL, testCase.requestedUserId)
			if err != nil && !testCase.wantError {
				t.Error("error should be nil")
			}

			if !(testCase.expectedPrincipal == nil && returnedPrincipal == nil) && !(testCase.expectedPrincipal != nil && returnedPrincipal != nil && testCase.expectedPrincipal.Id == returnedPrincipal.Id) {
				t.Errorf("returnedPrincipal '%v' is not equal to expected Principal '%v'", testCase.expectedPrincipal, returnedPrincipal)
			}
		})
	}
}
func newIdpStub(foundPrincipal *scim.Principal) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/identityprovider/scim/users/") {
			requestedPrincipalId := r.URL.Path[29:len(r.URL.Path)]
			if foundPrincipal != nil && foundPrincipal.Id == requestedPrincipalId {
				_ = json.NewEncoder(w).Encode(foundPrincipal)
			} else {
				http.Error(w, "", http.StatusNotFound)
			}
			return
		}
		http.Error(w, "", http.StatusNotFound)
	}))
}
