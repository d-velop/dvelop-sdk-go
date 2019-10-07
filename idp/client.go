// Package idp contains functions for the interaction with the IdentityProvider-App
package idp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/d-velop/dvelop-sdk-go/idp/scim"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var httpClient = &http.Client{}

var userCache = cache.New(1*time.Minute, 5*time.Minute)

var maxAgeRegex = regexp.MustCompile(`(?i)max-age=([^,\s]*)`) // cf. https://regex101.com/

func Validate(ctx context.Context, systemBaseUriString string, authSessionId string, tenantId string, loginfo func(ctx context.Context, logmessage string), allowExternalValidation bool) (scim.Principal, error) {
	cacheKey := fmt.Sprintf("%v/%v", tenantId, authSessionId)
	co, found := userCache.Get(cacheKey)
	if found {
		p := co.(scim.Principal)
		loginfo(ctx, fmt.Sprintf("taking user info for user '%v' from in memory cache.\n", p.Id))
		return p, nil
	}

	validateEndpoint, vEErr := validateEndpointFor(systemBaseUriString, allowExternalValidation)
	if vEErr != nil {
		return scim.Principal{}, vEErr
	}

	req, nRErr := http.NewRequest("GET", validateEndpoint.String(), nil)
	if nRErr != nil {
		return scim.Principal{}, fmt.Errorf("can't create http request for '%v' because: %v", validateEndpoint.String(), nRErr)
	}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	response, doErr := httpClient.Do(req)
	if doErr != nil {
		return scim.Principal{}, fmt.Errorf("error calling http GET on '%v' because: %v", validateEndpoint.String(), doErr)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		var p scim.Principal
		decErr := json.NewDecoder(response.Body).Decode(&p)
		if decErr != nil {
			return scim.Principal{}, fmt.Errorf("response from Identityprovider '%v' is no valid JSON because: %v", validateEndpoint.String(), decErr)
		}
		if p.Id == "" && !isPrincipalExternalUser(p) {
			return scim.Principal{}, errors.New("principal returned by identityprovider has no Id")
		}

		var validFor time.Duration = 0
		cacheControlHeader := response.Header.Get("Cache-Control")
		matches := maxAgeRegex.FindStringSubmatch(cacheControlHeader)
		if matches != nil {
			d, err := time.ParseDuration(matches[1] + "s")
			if err == nil {
				validFor = d
			}
		}
		if validFor > 0 {
			userCache.Set(cacheKey, p, validFor)
		}
		return p, nil
	case http.StatusUnauthorized:
		_, _ = ioutil.ReadAll(response.Body) // client must read to EOF and close body cf. https://godoc.org/net/http#Client
		return scim.Principal{}, ErrInvalidAuthSessionId
	case http.StatusForbidden:
		_, _ = ioutil.ReadAll(response.Body) // client must read to EOF and close body cf. https://godoc.org/net/http#Client
		return scim.Principal{}, ErrExternalValidationNotAllowed
	default:
		responseMsg, err := ioutil.ReadAll(response.Body)
		responseString := ""
		if err == nil {
			responseString = string(responseMsg)
		}
		return scim.Principal{}, fmt.Errorf(fmt.Sprintf("Identityprovider '%v' returned HTTP-Statuscode '%v' and message '%v'",
			response.Request.URL, response.StatusCode, responseString))
	}
}

func isPrincipalExternalUser(p scim.Principal) bool {
	for _, group := range p.Groups {
		if strings.ToUpper(group.Value) == "3E093BE5-CCCE-435D-99F8-544656B98681" {
			return true
		}
	}
	return false
}

func validateEndpointFor(systemBaseUriString string, allowExternalValidation bool) (*url.URL, error) {
	validateEndpointString := "/identityprovider/validate"
	if allowExternalValidation {
		validateEndpointString = fmt.Sprintf("%v?allowExternalValidation=true", validateEndpointString)
	}
	validateEndpoint, vPErr := url.Parse(validateEndpointString)
	if vPErr != nil {
		return nil, fmt.Errorf("%v", vPErr)
	}
	base, sBPErr := url.Parse(systemBaseUriString)
	if sBPErr != nil {
		return nil, fmt.Errorf("invalid SystemBaseUri '%v' because: %v", systemBaseUriString, sBPErr)
	}
	return base.ResolveReference(validateEndpoint), nil
}

var ErrInvalidAuthSessionId = errors.New("invalid AuthSessionId")

var ErrExternalValidationNotAllowed = errors.New("external validation not allowed")
