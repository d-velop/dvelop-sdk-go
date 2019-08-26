package lambdaenvironment_test

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/d-velop/dvelop-sdk-go/lambdaenvironment"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestToNamedAlias_ReturnsAliasFromArn(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name:some-version-tag",
	})
	handlerSpy := handlerSpy{}

	lambdaenvironment.AddEnvironmentToCtx()(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "some-version-tag" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, "some-version-tag")
	}
}

func TestRequestToVersionNumber_ReturnsVersionFromArn(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name:4711",
	})
	handlerSpy := handlerSpy{}

	lambdaenvironment.AddEnvironmentToCtx()(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "4711" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, "4711")
	}
}

func TestRequestToArnWithoutQualifier_ReturnsEmptyString(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name",
	})
	handlerSpy := handlerSpy{}

	lambdaenvironment.AddEnvironmentToCtx()(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, "")
	}
}

func TestRequestWithoutLambdaContext_ReturnsEmptyString(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	handlerSpy := handlerSpy{}

	lambdaenvironment.AddEnvironmentToCtx()(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, "")
	}
}

type handlerSpy struct {
	hasBeenCalled bool
	environment   string
}

func (spy *handlerSpy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	spy.hasBeenCalled = true
	spy.environment = lambdaenvironment.EnvironmentFromCtx(r.Context())
}
