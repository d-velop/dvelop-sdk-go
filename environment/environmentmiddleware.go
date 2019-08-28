package environment

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"net/http"
	"strings"
)

type contextKey string

const environmentKey = contextKey("FeatureToggleEnvironment")

type GetEnvironmentFromRequestFunc func(http.Request) string

// AddToCtx retrieves the current lambda alias/version and adds it to the context.
//
// You can use this to implement feature toggles that are dependent on the environment.
//
// Example:
//  func main() {
//    environmentFunc := func(request http.Request) string {
//    	if strings.HasPrefix(request.URL.Host, "dev.") {
//    	  return "dev"
//      } else {
//    	  return "prod"
//      }
//    }
//    mux := http.NewServeMux()
//    mux.Handle("/hello", environment.AddToCtx(environmentFunc)(someHandler()))
//  }
//
//  func someHandler() http.Handler {
//    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//      environment := environment.Get(r.Context())
//      if environment == "dev" {
//        fmt.Fprint(w, "Hey, here are some new features")
//      } else {
//        fmt.Fprint(w, "Hello, this is the production version")
//      }
//    })
//  }
//
// Or if you are running in lambda and want to use your lambda aliases:
//  func main() {
//    mux := http.NewServeMux()
//    mux.Handle("/hello", environment.AddToCtx(environment.FromLambdaContext)(someHandler()))
//  }
func AddToCtx(getEnvironmentFromRequest GetEnvironmentFromRequestFunc) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			environmentFromRequest := getEnvironmentFromRequest(*req)
			ctx = context.WithValue(ctx, environmentKey, environmentFromRequest)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func Get(ctx context.Context) string {
	value, ok := ctx.Value(environmentKey).(string)
	if !ok {
		return ""
	}
	return value
}

func FromLambdaContext(req http.Request) string {
	ctx := req.Context()

	if lc, success := lambdacontext.FromContext(ctx); success {
		arn := lc.InvokedFunctionArn
		arnParts := strings.Split(arn, ":")
		if len(arnParts) == 8 {
			return arnParts[7]
		}
	}

	return ""
}