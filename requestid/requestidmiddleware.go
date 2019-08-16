// Package requestid contains functions to handle an id for the current request.
//
// The idea is to read the current id from the received request and pass it
// to downstream services in order to get a trace across service boundaries.
//
// Logstatements should log the request id in every statement to correlate
// the statements to a specific request. This simplifies the tracking of
// a request through a system which serves multiple concurrent requests.
//
// Example:
//	func main() {
//		mux := http.NewServeMux()
//		mux.Handle("/hello", requestid.AddToCtx()(helloHandler()))
//	}
//
//	func helloHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			// get id from context
//			rid,_ := requestid.FromCtx(r.Context())
//		})
//	}
package requestid

import (
	"context"
	"errors"
	"net/http"

	"github.com/satori/go.uuid"
)

type contextKey string

const reqIdCtxKey = contextKey("reqId")
const reqIdHeader = "x-dv-request-id"

// AddToCtx reads the requestid http header x-dv-request-id from the current request
// and stores the id in the context.
//
// If the request doesn't have an existing id a new one is generated.
func AddToCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			reqId := req.Header.Get(reqIdHeader)
			if reqId == "" {
				reqId = uuid.Must(uuid.NewV4()).String()
			}
			ctx = context.WithValue(ctx, reqIdCtxKey, reqId)

			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

// FromCtx reads the current request id from the context.
func FromCtx(ctx context.Context) (string, error) {
	reqId, ok := ctx.Value(reqIdCtxKey).(string)
	if !ok {
		return "", errors.New("no requestid on context")
	}
	return reqId, nil
}
