// Package idp contains functions for the authentication process with the IdentityProviderApp
package idp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
	"github.com/patrickmn/go-cache"
)

type contextKey string

const principalKey = contextKey("Principal")
const authSessionIdKey = contextKey("AuthSessionId")

// HandleAuth authenticates the user using the IdentityProviderApp
//
// If the user is already logged in the credentials of the user are taken from the http request.
// Otherwise the request is redirected to the IdentityProvider for authentication and redirected back to
// the resource which has been originally invoked.
// If the user is logged in successfully information about the user (principal) and the authSession can be
// taken from the context.
// The parameter allowExternalValidation determines if the handler accepts external users. External users
// are those who have been successfully authenticated by an external identity provider such as Google, but
// have NOT been added to the pool of known users of this particular d.velop cloud tenant so far.
// USE THIS FEATURE WITH CAUTION. You don't know much about external users and should restrict the rights
// of external users to a minimum if you must allow access for external users at all. If external users are
// enabled (allowExternalValidation is true) the principal struct representing an external user doesn't
// provide any information apart from the e-mail address and a the reserved group ID '3E093BE5-CCCE-435D-99F8-544656B98681'
// which marks the user as an external user which is unknown to the system. This group ID can be used to
// distinguish external from internal users.
// If you are unsure, you should set allowExternalValidation to false, as you usually don't want external users to access your app.
//
// Example:
//	func main() {
//		// allow user which is authenticated by Open ID Connect provider
//		allowExternalValidation := true
//		mux := http.NewServeMux()
//		mux.Handle("/hello", idp.HandleAuth(tenant.SystemBaseUriFromCtx, tenant.IdFromCtx, allowExternalValidation, logerror, loginfo)(helloHandler()))
//	}
//
//	func helloHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			// get user from context
//			principal,_ := idp.PrincipalFromCtx(r.Context())
//			// get authSessionId From context
//			authSessionId,_ := idp.AuthSessionIdFromCtx(r.Context())
//			fmt.Fprintf(w, "Hello %v your authsessionId is %v", principal.DisplayName, authSessionId)
//		})
//	}
func HandleAuth(getSystemBaseUriFromCtx, getTenantIdFromCtx func(ctx context.Context) (string, error), allowExternalValidation bool, logerror, loginfo func(ctx context.Context, logmessage string)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			authSessionId, aErr := authSessionIdFrom(ctx, req, loginfo)
			if aErr != nil {
				logerror(ctx, fmt.Sprintf("error reading authSessionId from request because: %v\n", aErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if authSessionId == "" {
				acceptHeader := req.Header.Get("Accept")
				if !isTextHtmlAccepted(acceptHeader) {
					rw.WriteHeader(http.StatusUnauthorized)
					rw.Header().Set("WWW-Authenticate", "Bearer")
					return
				}

				switch req.Method {
				case "POST", "PUT", "DELETE", "PATCH":
					rw.WriteHeader(http.StatusUnauthorized)
					rw.Header().Set("WWW-Authenticate", "Bearer")
				default:
					redirectToIdpLogin(rw, req)
				}
				return
			}
			systemBaseUri, gSBErr := getSystemBaseUriFromCtx(ctx)
			if gSBErr != nil {
				logerror(ctx, fmt.Sprintf("error reading SystemBaseUri from context because: %v\n", gSBErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			tenantId, gTErr := getTenantIdFromCtx(req.Context())
			if gTErr != nil {
				logerror(ctx, fmt.Sprintf("error reading TenandId from context because: %v\n", gTErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			principal, gPErr := getPrincipalFromIdp(ctx, systemBaseUri, authSessionId, tenantId, loginfo, allowExternalValidation)
			if gPErr != nil {
				if gPErr == errInvalidAuthSessionId {
					redirectToIdpLogin(rw, req)
				} else if gPErr == errExternalValidationNotAllowed {
					loginfo(ctx, fmt.Sprintf("external user tries to access a resource and doesn't have sufficient rights."))
					http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				} else {
					logerror(ctx, fmt.Sprintf("error getting principal from Identityprovider because: %v\n", gPErr))
					http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				return
			}
			ctx = context.WithValue(ctx, authSessionIdKey, authSessionId)
			ctx = context.WithValue(ctx, principalKey, principal)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

func redirectToIdpLogin(rw http.ResponseWriter, req *http.Request) {
	const redirectionBase = "/identityprovider/login?redirect="
	rw.Header().Set("Location", redirectionBase+url.QueryEscape(req.URL.String()))
	rw.WriteHeader(http.StatusFound)
}

var httpClient = &http.Client{}

var errInvalidAuthSessionId = errors.New("invalid AuthSessionId")

var errExternalValidationNotAllowed = errors.New("external validation not allowed")

var userCache = cache.New(1*time.Minute, 5*time.Minute)

var maxAgeRegex = regexp.MustCompile(`(?i)max-age=([^,\s]*)`) // cf. https://regex101.com/

func isTextHtmlAccepted(header string) bool {

	trimmedHeader := strings.TrimSpace(header)
	if trimmedHeader == "" {
		return true
	}

	acceptableTypes := strings.Split(trimmedHeader, ",")
	for _, a := range acceptableTypes {
		parts := strings.SplitN(a, ";", 2)
		t := strings.TrimSpace(parts[0])

		if t == "*/*" || t == "text/*" {
			t = "text/html"
		}

		q := 1.0
		if len(parts) == 2 && len(parts[1]) > 2 {
			qPart := strings.TrimSpace(parts[1][3:])
			var err error
			q, err = strconv.ParseFloat(qPart, 64)
			if err != nil {
				q = 0
			}
		}

		if (t == "text/html") && (q > 0.0) {
			return true
		}
	}

	return false
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

func getPrincipalFromIdp(ctx context.Context, systemBaseUriString string, authSessionId string, tenantId string, loginfo func(ctx context.Context, logmessage string), allowExternalValidation bool) (scim.Principal, error) {
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
		return scim.Principal{}, errInvalidAuthSessionId
	case http.StatusForbidden:
		return scim.Principal{}, errExternalValidationNotAllowed
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

var bearerTokenRegex = regexp.MustCompile("^(?i)bearer (.*)$") // cf. https://regex101.com/

func authSessionIdFrom(ctx context.Context, req *http.Request, loginfo func(ctx context.Context, logmessage string)) (string, error) {
	authorizationHeader := req.Header.Get("Authorization")
	matches := bearerTokenRegex.FindStringSubmatch(authorizationHeader)
	if matches != nil {
		return matches[1], nil
	}

	const authSessionId = "AuthSessionId"
	for _, cookie := range req.Cookies() {
		if cookie.Name == authSessionId {
			// cookie is URL encoded cf. https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
			value, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				return "", fmt.Errorf("value '%v' of '%v'-cookie is no valid url escaped string because: %v", cookie.Value, cookie.Name, err)
			}
			return value, nil
		}
	}
	loginfo(ctx, fmt.Sprintf("no AuthSessionId found because there is no bearer authorization header and no AuthSessionId Cookie\n"))
	return "", nil
}

func PrincipalFromCtx(ctx context.Context) (scim.Principal, error) {
	principal, ok := ctx.Value(principalKey).(scim.Principal)
	if !ok {
		return scim.Principal{}, errors.New("no principal on context")
	}
	return principal, nil
}

func AuthSessionIdFromCtx(ctx context.Context) (string, error) {
	authSessionId, ok := ctx.Value(authSessionIdKey).(string)
	if !ok {
		return "", errors.New("no authSessionId on context")
	}
	return authSessionId, nil
}
