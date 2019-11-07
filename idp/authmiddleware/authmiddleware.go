// Package authmiddleware contains a http middleware for the authentication with the IdentityProvider-App
package authmiddleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

type contextKey string

const principalKey = contextKey("Principal")
const authSessionIdKey = contextKey("AuthSessionId")

// Authenticate authenticates the user using the IdentityProvider-App
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
//		idpClient := &idp.Client{}
//		authenticate := authmiddleware.Authenticate(idpClient, tenant.SystemBaseUriFromCtx, tenant.IdFromCtx, false, logerror, loginfo)
//		authenticateExternal := authmiddleware.Authenticate(idpClient, tenant.SystemBaseUriFromCtx, tenant.IdFromCtx, true, logerror, loginfo)
//		mux := http.NewServeMux()
//		mux.Handle("/hello", authenticate(helloHandler()))
//		mux.Handle("/resource", authenticate(resourceHandler()))
//		mux.Handle("/resource4ExternalUsers", authenticateExternal(resource4ExternalUsersHandler()))
//	}
//
//	func helloHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			// get user from context
//			principal,err := authmiddleware.PrincipalFromCtx(r.Context())
//			// get authSessionId From context
//			authSessionId,err := authmiddleware.AuthSessionIdFromCtx(r.Context())
//			fmt.Fprintf(w, "Hello %v your authsessionId is %v", principal.DisplayName, authSessionId)
//		})
//	}
func Authenticate(validator Validator, getSystemBaseUriFromCtx, getTenantIdFromCtx func(ctx context.Context) (string, error), allowExternalValidation bool, logError, logInfo func(ctx context.Context, message string)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			authSessionId, aErr := authSessionIdFromRequest(ctx, req, logInfo)
			if aErr != nil {
				logError(ctx, fmt.Sprintf("error reading authSessionId from request because: %v\n", aErr))
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
				logError(ctx, fmt.Sprintf("error reading SystemBaseUri from context because: %v\n", gSBErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			tenantId, gTErr := getTenantIdFromCtx(ctx)
			if gTErr != nil {
				logError(ctx, fmt.Sprintf("error reading TenandId from context because: %v\n", gTErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			principal, valErr := validator.Validate(ctx, systemBaseUri, tenantId, authSessionId)
			if valErr != nil {
				logError(ctx, fmt.Sprintf("error getting principal from Identityprovider because: %v\n", valErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if principal == nil {
				redirectToIdpLogin(rw, req)
				return
			}
			if principal.IsExternal() && !allowExternalValidation {
				logInfo(ctx, fmt.Sprintf("external user tries to access a resource and doesn't have sufficient rights."))
				http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			ctx = context.WithValue(ctx, authSessionIdKey, authSessionId)
			ctx = context.WithValue(ctx, principalKey, *principal)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

// Validator is an interface representing the ability to validate an authSessionId
type Validator interface {
	// Validate checks if the authSessionId is valid for the tenant specified by systemBaseUri and tenantId.
	//
	// If the authSessionId is valid, that is it belongs to a principal and has not expired, a none nil *scim.Principal is returned.
	// Otherwise the returned *scim.Principal is nil.
	//
	// An error is returned if something unexpected occurred.
	Validate(ctx context.Context, systemBaseUri string, tenantId string, authSessionId string) (*scim.Principal, error)
}

func redirectToIdpLogin(rw http.ResponseWriter, req *http.Request) {
	const redirectionBase = "/identityprovider/login?redirect="
	rw.Header().Set("Location", redirectionBase+url.QueryEscape(req.URL.String()))
	rw.WriteHeader(http.StatusFound)
}

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

var bearerTokenRegex = regexp.MustCompile("^(?i)bearer (.*)$") // cf. https://regex101.com/

func authSessionIdFromRequest(ctx context.Context, req *http.Request, logInfo func(ctx context.Context, message string)) (string, error) {
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
				return "", fmt.Errorf("value '%s' of '%s'-cookie is no valid url escaped string because: %v", cookie.Value, cookie.Name, err)
			}
			return value, nil
		}
	}
	logInfo(ctx, fmt.Sprintf("no AuthSessionId found because there is no bearer authorization header and no AuthSessionId Cookie\n"))
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
