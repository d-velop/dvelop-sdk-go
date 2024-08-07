package lambda

import (
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

func GetAliasFromRequest(req http.Request) string {
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
