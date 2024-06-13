package lambda_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/d-velop/dvelop-sdk-go/lambda"
)

func TestGetAliasFromRequest_LambdaArnHasNamedLambdaAlias_ReturnsAliasFromArn(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name:some-version-tag",
	})

	environment := lambda.GetAliasFromRequest(*req.WithContext(ctx))

	if environment != "some-version-tag" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", environment, "some-version-tag")
	}
}

func TestGetAliasFromRequest_LambdaArnHasVersionNumber_ReturnsVersionFromArn(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name:4711",
	})

	environment := lambda.GetAliasFromRequest(*req.WithContext(ctx))

	if environment != "4711" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", environment, "4711")
	}
}

func TestGetAliasFromRequest_LambdaArnHasNoQualifier_ReturnsEmptyString(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name",
	})

	environment := lambda.GetAliasFromRequest(*req.WithContext(ctx))

	if environment != "" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", environment, "")
	}
}

func TestGetAliasFromRequest_RequestWithoutLambdaContext_ReturnsEmptyString(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	environment := lambda.GetAliasFromRequest(*req)

	if environment != "" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", environment, "")
	}
}
