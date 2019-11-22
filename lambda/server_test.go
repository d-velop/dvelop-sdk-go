package lambda_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/d-velop/dvelop-sdk-go/lambda"
)

func invokeAdaptorFunc(t *testing.T, evt *events.APIGatewayProxyRequest) *testresult {
	spy := &handlerSpy{}
	adaptorFunc := lambda.AdaptorFunc(spy, nullLog, nullLog)
	_, _ = adaptorFunc(context.Background(), *evt)
	return &testresult{t: t, input: evt, req: spy.req}
}

func nullLog(c context.Context, s string) {
	_ = c
	_ = s
}

type handlerSpy struct {
	req         *http.Request
	handlerFunc func(http.ResponseWriter, *http.Request)
}

func (hs *handlerSpy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hs.req = r
	if hs.handlerFunc != nil {
		hs.handlerFunc(w, r)
	}
}

type testresult struct {
	t     *testing.T
	input *events.APIGatewayProxyRequest
	req   *http.Request
}

func TestAdaptor_InvokesHandlerWithCorrectMethod(t *testing.T) {
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "Get"}).invokesHandlerWithMethod(http.MethodGet)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "get"}).invokesHandlerWithMethod(http.MethodGet)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "GET"}).invokesHandlerWithMethod(http.MethodGet)

	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "POST"}).invokesHandlerWithMethod(http.MethodPost)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "HEAD"}).invokesHandlerWithMethod(http.MethodHead)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "PUT"}).invokesHandlerWithMethod(http.MethodPut)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "PATCH"}).invokesHandlerWithMethod(http.MethodPatch)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "DELETE"}).invokesHandlerWithMethod(http.MethodDelete)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "CONNECT"}).invokesHandlerWithMethod(http.MethodConnect)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "OPTIONS"}).invokesHandlerWithMethod(http.MethodOptions)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{HTTPMethod: "TRACE"}).invokesHandlerWithMethod(http.MethodTrace)
}

func (tr *testresult) invokesHandlerWithMethod(expected string) {
	if tr.req == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with request.method '%v' but request was nil ", tr.input, expected)
	}

	if tr.req.Method != expected {
		tr.t.Errorf("Serve(%v): should invoke handler with request.method '%v' but request.method was '%v' ", tr.input, expected, tr.req.Method)
	}
}

func TestAdaptor_InvokesHandlerWithCorrectURL(t *testing.T) {
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Path: "/path"}).invokesHandlerWithURL(&url.URL{Path: "/path"})
	// sort query parameter by key
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Path: "/path", QueryStringParameters: map[string]string{"bar": "2", "foo": "1"}}).invokesHandlerWithURL(&url.URL{Path: "/path", RawQuery: "bar=2&foo=1"})
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Path: "/path", QueryStringParameters: map[string]string{"foo": "1", "bar": "2"}}).invokesHandlerWithURL(&url.URL{Path: "/path", RawQuery: "bar=2&foo=1"})
}

func (tr *testresult) invokesHandlerWithURL(expected *url.URL) {
	if tr.req == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with request.URL '%v' but request was nil ", tr.input, expected)
	}

	if !reflect.DeepEqual(tr.req.URL, expected) {
		tr.t.Errorf("Serve(%v): should invoke handler with request.URL '%v' but request.URL was '%v' ", tr.input, expected, tr.req.URL)
	}
}

func TestAdaptor_InvokesHandlerWithCorrectHeader(t *testing.T) {
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Headers: nil}).invokesHandlerWithHeader(http.Header{})
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Headers: map[string]string{}}).invokesHandlerWithHeader(http.Header{})

	expected := http.Header{}
	expected.Add("Accept", "application/json")
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Headers: map[string]string{"Accept": "application/json"}}).invokesHandlerWithHeader(expected)
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Headers: map[string]string{"accept": "application/json"}}).invokesHandlerWithHeader(expected)

	expected = http.Header{}
	expected.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Headers: map[string]string{"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"}}).invokesHandlerWithHeader(expected)

	expected = http.Header{}
	expected.Add("Accept-Encoding", "gzip, deflate")
	expected.Add("Accept-Language", "en-us")
	expected.Add("Foo", "Bar")
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Headers: map[string]string{
		"accept-encoding": "gzip, deflate",
		"Accept-Language": "en-us",
		"fOO":             "Bar"}}).invokesHandlerWithHeader(expected)
}

func (tr *testresult) invokesHandlerWithHeader(expected http.Header) {
	if tr.req == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with request.Header '%v' but request was nil", tr.input, expected)
	}

	if !reflect.DeepEqual(tr.req.Header, expected) {
		tr.t.Errorf("Serve(%v): should invoke handler with request.Header '%v' but request.Header was '%v' ", tr.input, expected, tr.req.Header)
	}
}

func TestAdaptor_InvokesHandlerWithCorrectRequestURI(t *testing.T) {
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Path: "/path"}).invokesHandlerWithRequestURI("/path")
	// sort query parameter by key
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Path: "/path", QueryStringParameters: map[string]string{"foo": "1", "bar": "2"}}).invokesHandlerWithRequestURI("/path?bar=2&foo=1")
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Path: "/path", QueryStringParameters: map[string]string{"bar": "2", "foo": "1"}}).invokesHandlerWithRequestURI("/path?bar=2&foo=1")
}

func (tr *testresult) invokesHandlerWithRequestURI(expected string) {
	if tr.req == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with request.RequestURI '%v' but request was nil", tr.input, expected)
	}

	if !reflect.DeepEqual(tr.req.RequestURI, expected) {
		tr.t.Errorf("Serve(%v): should invoke handler with request.RequestURI '%v' but request.RequestURI was '%v' ", tr.input, expected, tr.req.RequestURI)
	}
}

func TestAdaptor_InvokesHandlerWithCorrectBody(t *testing.T) {
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Body: "", IsBase64Encoded: false}).invokesHandlerWithBody("")
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Body: "Hallo Welt", IsBase64Encoded: false}).invokesHandlerWithBody("Hallo Welt")
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Body: "<h1>Hallo Welt</h1>", IsBase64Encoded: false}).invokesHandlerWithBody("<h1>Hallo Welt</h1>")
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Body: `{"Herausgeber": "Xema","Inhaber":{"Name": "Mustermann","maennlich": true,"Alter": 42}}`, IsBase64Encoded: false}).
		invokesHandlerWithBody(`{"Herausgeber": "Xema","Inhaber":{"Name": "Mustermann","maennlich": true,"Alter": 42}}`)

	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Body: "", IsBase64Encoded: true}).invokesHandlerWithBase64Body([]byte{})
	invokeAdaptorFunc(t, &events.APIGatewayProxyRequest{Body: base64.StdEncoding.EncodeToString([]byte("Hallo Welt")), IsBase64Encoded: true}).invokesHandlerWithBase64Body([]byte("Hallo Welt"))
}

func (tr *testresult) invokesHandlerWithBody(expected string) {
	if tr.req == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with request.Body '%v' but request was nil", tr.input, expected)
	}

	// cf https://godoc.org/net/http#Request:
	if tr.req.Body == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with none-nil request.Body '%v' but request.Body was nil ", tr.input, expected)
	}

	if expected == "" {
		var b []byte
		_, e := tr.req.Body.Read(b)
		if e != io.EOF {
			tr.t.Errorf("Serve(%v): should invoke handler with request.Body which immediately returns EOF error but returned error was '%v'", tr.input, e)
		}
	} else {
		b, err := ioutil.ReadAll(tr.req.Body)
		if err != nil {
			tr.t.Fatalf("Serve(%v): should invoke handler with a valid request.Body but got an error '%v' while reading the request.Body", tr.input, err)
		}
		bs := string(b)
		if bs != expected {
			tr.t.Errorf("Serve(%v): should invoke handler with Request.Body '%v' but request.Body was '%v' ", tr.input, expected, bs)
		}
	}
}

func (tr *testresult) invokesHandlerWithBase64Body(expected []byte) {
	if tr.req == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with request.Body '%v' but request was nil", tr.input, expected)
	}

	// cf https://godoc.org/net/http#Request:
	if tr.req.Body == nil {
		tr.t.Fatalf("Serve(%v): should invoke handler with none-nil request.Body '%v' but request.Body was nil ", tr.input, expected)
	}

	if len(expected) == 0 {
		var b []byte
		_, e := tr.req.Body.Read(b)
		if e != io.EOF {
			tr.t.Errorf("Serve(%v): should invoke handler with request.Body which immediately returns EOF error but returned error was '%v'", tr.input, e)
		}
	} else {
		b, err := ioutil.ReadAll(tr.req.Body)
		if err != nil {
			tr.t.Fatalf("Serve(%v): should invoke handler with a valid request.Body but got an error '%v' while reading the request.Body", tr.input, err)
		}
		if !bytes.Equal(b, expected) {
			tr.t.Errorf("Serve(%v): should invoke handler with Request.Body '%v' but request.Body was '%v' ", tr.input, expected, b)
		}
	}
}

func TestAdaptor_HandlerDoesNothing_ReturnsEmptyBodyAndNoHeadersAndStatusOK(t *testing.T) {
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	if resp.Body != "" {
		t.Errorf("Serve: should return empty body but returned '%v'", resp.Headers)
	}

	if resp.Headers != nil {
		t.Errorf("Serve: should return nil headers but returned '%v'", resp.Headers)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Serve: should status '%v' but returned '%v'", http.StatusOK, resp.StatusCode)
	}
}

func TestAdaptor_HandlerSetsHeaderAndCallsWriteHeader_ReturnsHeaderAndStatusCode(t *testing.T) {
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Error", "502")
		w.WriteHeader(http.StatusInternalServerError)
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	expected := map[string]string{"X-Error": "502"}
	if !reflect.DeepEqual(resp.Headers, expected) {
		t.Errorf("Serve: should return headers '%v' set by handler but returned headers '%v' ", expected, resp.Headers)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Serve: should return StatusCode '%v' set by handler but returned StatusCode '%v' ", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestAdaptor_HandlerSetsHeaderWithMultipleValues_ReturnsHeaderWithMultipleValuesAndStatusCode(t *testing.T) {
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	expected := map[string]string{"Content-Type": "application/json, text/html"}
	if !reflect.DeepEqual(resp.Headers, expected) {
		t.Errorf("Serve: should return headers '%v' set by handler but returned headers '%v' ", expected, resp.Headers)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Serve: should return StatusCode '%v' set by handler but returned StatusCode '%v' ", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestAdaptor_HandlerModifiesHeaderAfterCallingWriteHeader_ReturnsUnmodifiedHeaders(t *testing.T) {
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Header", "value")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Header().Add("Content-Type", "text/html")
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	expected := map[string]string{"X-Header": "value"}
	if !reflect.DeepEqual(expected, resp.Headers) {
		t.Errorf("Serve: should return headers '%v' set by handler but returned headers '%v' ", expected, resp.Headers)
	}
}

func TestAdaptor_HandlerDoesntSetContentTypeAndCallsWrite_ReturnsDetectedContentTypeAndBodyAndStatusCodeOK(t *testing.T) {
	body := "<h1>Hello world!</h1>"
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, body)
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	expected := map[string]string{"Content-Type": "text/html; charset=utf-8"}
	if !reflect.DeepEqual(resp.Headers, expected) {
		t.Errorf("Serve: should return headers '%v' set by handler but returned headers '%v' ", expected, resp.Headers)
	}
	if !reflect.DeepEqual(resp.Body, body) {
		t.Errorf("Serve: should return body '%v' set by handler but returned body '%v' ", body, resp.Body)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Serve: should return StatusCode '%v' set by handler but returned StatusCode '%v' ", http.StatusOK, resp.StatusCode)
	}
}

func TestAdaptor_HandlerSetsContentTypeAndCallsWrite_ReturnsContentTypeHeaderAndBodyAndStatusCodeOK(t *testing.T) {
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd+company.category+json")
		_, _ = fmt.Fprintf(w, `{"Key": "value"}`)
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	expected := map[string]string{"Content-Type": "application/vnd+company.category+json"}
	if !reflect.DeepEqual(resp.Headers, expected) {
		t.Errorf("Serve: should return headers '%v' set by handler but returned headers '%v' ", expected, resp.Headers)
	}
	if !reflect.DeepEqual(resp.Body, `{"Key": "value"}`) {
		t.Errorf("Serve: should return body '%v' set by handler but returned body '%v' ", `{"Key": "value"}`, resp.Body)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Serve: should return StatusCode '%v' set by handler but returned StatusCode '%v' ", http.StatusOK, resp.StatusCode)
	}
}

func TestAdaptor_HandlerSetsContentTypeAndCallsWriteHeaderAndCallsWrite_ReturnsContentTypeHeaderAndBodyAndStatusCode(t *testing.T) {
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Header", "value")
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = fmt.Fprintf(w, "type not acceptable")
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	expected := map[string]string{"X-Header": "value"}
	if !reflect.DeepEqual(resp.Headers, expected) {
		t.Errorf("Serve: should return headers '%v' set by handler but returned headers '%v' ", expected, resp.Headers)
	}
	if !reflect.DeepEqual(resp.Body, "type not acceptable") {
		t.Errorf("Serve: should return body '%v' set by handler but returned body '%v' ", "type not acceptable", resp.Body)
	}
	if resp.StatusCode != http.StatusNotAcceptable {
		t.Errorf("Serve: should return StatusCode '%v' set by handler but returned StatusCode '%v' ", http.StatusNotAcceptable, resp.StatusCode)
	}
}

func TestAdaptor_HandlerModifiesHeaderAfterWriteAndCallsWriteHeader_ReturnsUnmodifiedHeaders(t *testing.T) {
	spy := &handlerSpy{handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Header", "value")
		_, _ = fmt.Fprintf(w, `{"Key": "value"}`)
		w.Header().Set("X-Header", "modifiedvalue")
		w.WriteHeader(http.StatusOK)
	}}

	handler := lambda.AdaptorFunc(spy, nullLog, nullLog)
	resp, _ := handler(context.Background(), events.APIGatewayProxyRequest{})

	if resp.Headers["X-Header"] != "value" {
		t.Errorf("Serve: should return unmodified header value '%v' but returned modified header value '%v'", "value", resp.Headers["X-Header"])
	}
}

var _, _ = lambda.ReqIdFromCtx(context.Background())
