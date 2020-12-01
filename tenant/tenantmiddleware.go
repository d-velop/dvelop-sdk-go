// Package tenant contains functions to determine for which tenant a request should be served.
//
// As soon as an App is registered in d.velop cloud requests for ALL tenants which booked the app
// are routed to the App. So each App MUST be able to serve requests for multiple tenants concurrently.
//
// Each request contains http headers (x-dv-tenant-id, x-dv-baseuri, x-dv-sig-1) which MUST be evaluated
// by the App to determine the tenant for which the request is meant.
//
// The tenant-id is meant as a stable and unique identifier for a tenant which doesn't change over time.
// It can be used to store tenant specific data.
//
// The systemBaseUri is the baseadress of the system for the particular tenant. It MUST be used if
// an App makes requests to other Apps on behalf of a tenant.
//
// This package contains functions which read the tenant information from a request and put them in the context.
//
// Example:
//	func main() {
//		mux := http.NewServeMux()
//		// NOTE: Each App gets it's own signature secret. So please change the following code accordingly and do not share your secret!
//		mux.Handle("/hello", tenant.AddToCtx(os.Getenv("systemBaseUri"), base64.StdEncoding.DecodeString("U2VjcmV0"))(helloHandler()))
//	}
//
//	func helloHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			// get systembaseuri from context
//			systembaseuri,_ := tenant.SystemBaseUriFromCtx(r.Context())
//			// get tenant id from context
//			tenant,_ := tenant.IdFromCtx(r.Context())
//		})
//	}
package tenant

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const (
	systemBaseUriCtxKey          = contextKey("systemBaseUri")
	tenantIdCtxKey               = contextKey("tenantId")
	initiatorSystemBaseUriCtxKey = contextKey("sourceSystemBaseUri")
	systemBaseUriHeader          = "x-dv-baseuri"
	tenantIdHeader               = "x-dv-tenant-id"
	forwardedHeader              = "forwarded"
	xForwardedHostHeader         = "x-forwarded-host"
	commaDelimiter               = ","
	colonDelimiter               = ";"
	forwardedHostPattern         = "host="
	uriPrefix                    = "https://"
)

// Adds systemBaseUri and tenantId to request context.
// If the headers are not present the given defaultSystemBaseUri and tenant "0" are used.
// The signatureSecretKey is specific for each App and is provided by the registration process for d.velop cloud.
func AddToCtx(defaultSystemBaseUri string, signatureSecretKey []byte) func(http.Handler) http.Handler {

	errorLogger := func(ctx context.Context, message string) {
		log.Println("error", message)
	}

	return AddToCtxWithLogger(defaultSystemBaseUri, signatureSecretKey, errorLogger)
}

func AddToCtxWithLogger(defaultSystemBaseUri string, signatureSecretKey []byte, logger func(ctx context.Context, message string)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			systemBaseUri := req.Header.Get(systemBaseUriHeader)
			tenantId := req.Header.Get(tenantIdHeader)

			if systemBaseUri != "" || tenantId != "" {
				if signatureSecretKey == nil {
					logger(req.Context(), fmt.Sprintf("validating signature for headers '%v' and '%v' because secret signature key has not been configured", systemBaseUriHeader, tenantIdHeader))
					http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				base64Signature := req.Header.Get("x-dv-sig-1")
				signature, err := base64.StdEncoding.DecodeString(base64Signature)
				if err != nil {
					logger(req.Context(), fmt.Sprintf("decoding signature '%v' as base 64 data because: %v", base64Signature, err))
					http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
				if !signatureIsValid([]byte(systemBaseUri+tenantId), []byte(signature), signatureSecretKey) {
					logger(req.Context(), fmt.Sprintf("signature '%v' is not valid for SystemBaseUri '%v' and TenantId '%v'", signature, systemBaseUri, tenantId))
					http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
			}

			if tenantId == "" {
				// tenant 0 is reserved for environments which don't support multitenancy and
				// therefore can not transmit tenant headers. So there is only one tenant "0".
				// As soon as this environment supports additonal tenants these additional tenants will
				// have an id != "0"
				tenantId = "0"
			}
			if tenantId != "" {
				ctx = context.WithValue(ctx, tenantIdCtxKey, tenantId)
			}

			if systemBaseUri == "" {
				systemBaseUri = defaultSystemBaseUri
			}
			if systemBaseUri != "" {
				ctx = context.WithValue(ctx, systemBaseUriCtxKey, systemBaseUri)
			}

			initiatorSystemBaseUri := getInitiatorSystemBaseUri(req)
			if initiatorSystemBaseUri == "" {
				initiatorSystemBaseUri = defaultSystemBaseUri
			}
			if initiatorSystemBaseUri != "" {
				ctx = context.WithValue(ctx, initiatorSystemBaseUriCtxKey, initiatorSystemBaseUri)
			}
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

func signatureIsValid(message, signature, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(signature, expectedMAC)
}

// returns the initial host which initiates current request
// it is essential in hybrid systems
func getInitiatorSystemBaseUri(req *http.Request) string {
	var initiatorSystemBaseUri string
	forwardedHeaderValue := req.Header.Get(forwardedHeader)
	xForwardedHostHeaderValue := req.Header.Get(xForwardedHostHeader)
	systemBaseUri := req.Header.Get(systemBaseUriHeader)

	initiatorSystemBaseUri = getForwardedHeaderFirstHostValueAsUri(forwardedHeaderValue)
	if initiatorSystemBaseUri == "" {
		host := getFirstValueOfDelimitedList(xForwardedHostHeaderValue, commaDelimiter)
		if host != "" {
			initiatorSystemBaseUri = uriPrefix + host
		} else {
			initiatorSystemBaseUri = systemBaseUri
		}
	}
	return initiatorSystemBaseUri
}

func getForwardedHeaderFirstHostValueAsUri(headerValue string) string {
	if headerValue != "" {
		for _, value := range strings.Split(headerValue, colonDelimiter) {
			if strings.HasPrefix(value, forwardedHostPattern) {
				hostValue := strings.TrimPrefix(value, forwardedHostPattern)
				host := getFirstValueOfDelimitedList(hostValue, commaDelimiter)
				if host != "" {
					return uriPrefix + host
				}
			}
		}
	}
	return ""
}

func getFirstValueOfDelimitedList(delimitedList string, delimiter string) string {
	if delimitedList == "" {
		return delimitedList
	}
	return strings.Split(delimitedList, delimiter)[0]
}

// SystemBaseUriFromCtx reads the systemBaseUri from the context.
func SystemBaseUriFromCtx(ctx context.Context) (string, error) {
	systemBaseUri, ok := ctx.Value(systemBaseUriCtxKey).(string)
	if !ok {
		return "", errors.New("no SystemBaseUri on context")
	}
	return systemBaseUri, nil
}

// IdFromCtx reads the tenant id from the context.
func IdFromCtx(ctx context.Context) (string, error) {
	tenantId, ok := ctx.Value(tenantIdCtxKey).(string)
	if !ok {
		return "", errors.New("no TenantId on context")
	}
	return tenantId, nil
}

// InitiatorSystemBaseUriFromCtx reads the uri of the initial requesting host from the context.
func InitiatorSystemBaseUriFromCtx(ctx context.Context) (string, error) {
	initiatorSystemBaseUri, ok := ctx.Value(initiatorSystemBaseUriCtxKey).(string)
	if !ok {
		return "", errors.New("no InitiatorSystemBaseUri on context")
	}
	return initiatorSystemBaseUri, nil
}

// SetId returns a new context.Context with the given tenantId
func SetId(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, tenantIdCtxKey, tenantId)
}

// SetSystemBaseUri returns a new context.Context with the given systemBaseUri
func SetSystemBaseUri(ctx context.Context, systemBaseUri string) context.Context {
	return context.WithValue(ctx, systemBaseUriCtxKey, systemBaseUri)
}

// SetInitiatorSystemBaseUri returns a new context.Context with the given initiatorSystemBaseUri
func SetInitiatorSystemBaseUri(ctx context.Context, initiatorSystemBaseUri string) context.Context {
	return context.WithValue(ctx, initiatorSystemBaseUriCtxKey, initiatorSystemBaseUri)
}
