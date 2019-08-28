package environment_test

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/d-velop/dvelop-sdk-go/lambdaenvironment"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestWithEnvironmentFunction_UsesReturnedEnvironment(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	handlerSpy := handlerSpy{}

	expectedEnvironment := "some-environment"

	environmentFunc := func(request http.Request) string {
		return expectedEnvironment
	}
	environment.AddToCtx(environmentFunc)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != expectedEnvironment {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, expectedEnvironment)
	}
}

func TestRequestToNamedLambdaAlias_ReturnsAliasFromArn(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name:some-version-tag",
	})
	handlerSpy := handlerSpy{}

	environment.AddToCtx(environment.FromLambdaContext)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "some-version-tag" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, "some-version-tag")
	}
}

func TestRequestToLambdaVersionNumber_ReturnsVersionFromArn(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name:4711",
	})
	handlerSpy := handlerSpy{}

	environment.AddToCtx(environment.FromLambdaContext)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "4711" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, "4711")
	}
}

func TestRequestToLambdaArnWithoutQualifier_ReturnsEmptyString(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-central-1:123456789012:function:some-function-name",
	})
	handlerSpy := handlerSpy{}

	environment.AddToCtx(environment.FromLambdaContext)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "" {
		t.Errorf("middleware returned wrong environment: got %v, want %v", handlerSpy.environment, "")
	}
}

func TestRequestForLambdaWithoutLambdaContext_ReturnsEmptyString(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	handlerSpy := handlerSpy{}

	environment.AddToCtx(environment.FromLambdaContext)(&handlerSpy).ServeHTTP(httptest.NewRecorder(), req)

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
	spy.environment = environment.Get(r.Context())
}
