package lambdaenvironment

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"net/http"
	"strings"
)

type contextKey string
const environmentKey = contextKey("FeatureToggleEnvironment")

func AddEnvironmentToCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
}

func EnvironmentFromCtx( ctx context.Context) string {
	value, ok := ctx.Value(environmentKey).(string)
	if !ok {
		return ""
	}
	return value
}