package environment_test

import (
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


func TestRequestWithoutMiddleware_GetterReturnsEmptyString(t *testing.T) {
	req, err := http.NewRequest("GET", "/somewhere", nil)
	if err != nil {
		t.Fatal(err)
	}

	handlerSpy := handlerSpy{}

	handlerSpy.ServeHTTP(httptest.NewRecorder(), req)

	if !handlerSpy.hasBeenCalled {
		t.Error("inner handler should have been called")
	}

	if handlerSpy.environment != "" {
		t.Errorf("middleware returned wrong environment: got '%v', want '%v'", handlerSpy.environment, "")
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
