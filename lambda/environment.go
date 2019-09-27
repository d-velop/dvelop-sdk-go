package lambda

import (
	"github.com/aws/aws-lambda-go/lambdacontext"
	"net/http"
	"strings"
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