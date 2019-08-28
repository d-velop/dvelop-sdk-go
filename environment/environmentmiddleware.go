package lambdaenvironment

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"net/http"
	"strings"
)

type contextKey string

const environmentKey = contextKey("FeatureToggleEnvironment")

type GetEnvironmentFromRequestFunc func(http.Request) string

// AddEnvironmentToCtx retrieves the current lambda alias/version and adds it to the context.
//
// You can use this to implement feature toggles that are dependent on the environment.
//
// Example:
//  func main() {
//    mux := http.NewServeMux()
//    mux.Handle("/hello", lambdaenvironment.AddEnvironmentToCtx(someHandler()))
//  }
//
//  func someHandler() http.Handler {
//    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//      environment := lambdaenvironment.EnvironmentFromCtx(r.Context())
//      if environment == "dev" {
//        fmt.Fprint(w, "Hey, here are some new features")
//      } else {
//        fmt.Fprint(w, "Hello, this is the production version")
//      }
//    })
//  }
func AddEnvironmentToCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if lc, success := lambdacontext.FromContext(ctx); success {
			arn := lc.InvokedFunctionArn
			arnParts := strings.Split(arn, ":")
			if len(arnParts) == 8 {
				ctx = context.WithValue(ctx, environmentKey, arnParts[7])
			}
		}

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func EnvironmentFromCtx(ctx context.Context) string {
	value, ok := ctx.Value(environmentKey).(string)
	if !ok {
		return ""
	}
	return value
}
