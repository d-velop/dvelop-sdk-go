// Package lambda provides a kind of http server emulator for lambda functions.
// That is an adapter which translates an incoming AWS APIGatewayProxyRequests to a regular http.Request and
// an outgoing http response to a AWS APIGatewayProxyResponse
//
// The use case are http applications which should be packaged and deployed as lambda functions and possibly at the same time
// as a normal local executable webapplication. So appart from this adapter function the rest of the application is unaware
// of the fact that it runs as a lambda function.
//
// So instead of
//	func main(){
//		//...
//		http.Serve (socket, handler)
//	}
// for a regular http server
//	func main(){
//		//...
//		lambda.Serve (handler, logerror, loginfo)
//	}
// can be used to serve http applications from lambda functions
package lambda

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// Serve uses a regular http.Handler to serve AWS APIGatewayProxyRequests
//
// Example:
//	func main(){
//		//...
//		lambda.Serve (handler, logerror, loginfo)
//	}
func Serve(handler http.Handler, logerror, loginfo func(ctx context.Context, logmessage string)) {
	lambda.Start(AdaptorFunc(handler, logerror, loginfo))
}

// AdaptorFunc adapts a regular http.Handler to an AWS lambda handler
func AdaptorFunc(handler http.Handler, logerror, loginfo func(ctx context.Context, logmessage string)) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fn := func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if lc, success := lambdacontext.FromContext(ctx); success {
			ctx = AddReqIdToCtx(ctx, lc.AwsRequestID)
		}
		loginfo(ctx, fmt.Sprintf("Received APIGatewayRequest '%v'", request.RequestContext.RequestID))
		respw := &responseWriter{header: http.Header{}, body: &bytes.Buffer{}}
		req, err := newRequest(&request)
		if err != nil {
			logerror(ctx, fmt.Sprint(err))
			resp := events.APIGatewayProxyResponse{Body: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError}
			return resp, nil
		}
		handler.ServeHTTP(respw, req.WithContext(ctx))
		resp, err := respw.response()
		if err != nil {
			logerror(ctx, fmt.Sprint(err))
			resp := events.APIGatewayProxyResponse{Body: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError}
			return resp, nil
		}
		return *resp, nil
	}
	return fn
}

func newRequest(evt *events.APIGatewayProxyRequest) (*http.Request, error) {
	req := &http.Request{
		Method: mapMethod(evt),
		URL:    mapURL(evt),
		Header: *mapHeader(evt),
	}

	if req.URL.RawQuery != "" {
		req.RequestURI = req.URL.Path + "?" + req.URL.RawQuery
	} else {
		req.RequestURI = req.URL.Path
	}

	if evt.IsBase64Encoded {
		decodedString, err := base64.StdEncoding.DecodeString(evt.Body)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Decoding of base64 body failed! cause:%v", err))
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(decodedString))
	} else {
		req.Body = ioutil.NopCloser(strings.NewReader(evt.Body))
	}
	return req, nil
}

func mapMethod(e *events.APIGatewayProxyRequest) string {
	switch strings.ToUpper(e.HTTPMethod) {
	case "GET":
		return http.MethodGet
	case "POST":
		return http.MethodPost
	case "HEAD":
		return http.MethodHead
	case "PUT":
		return http.MethodPut
	case "PATCH":
		return http.MethodPatch
	case "DELETE":
		return http.MethodDelete
	case "CONNECT":
		return http.MethodConnect
	case "OPTIONS":
		return http.MethodOptions
	case "TRACE":
		return http.MethodTrace
	default:
		return ""
	}
}

func mapURL(e *events.APIGatewayProxyRequest) *url.URL {
	u := &url.URL{Path: e.Path}
	if e.QueryStringParameters != nil {
		qspKeys := make([]string, 0, len(e.QueryStringParameters))
		for k := range e.QueryStringParameters {
			qspKeys = append(qspKeys, k)
		}
		sort.Strings(qspKeys)

		var buf bytes.Buffer
		var l = len(e.QueryStringParameters)

		for i, k := range qspKeys {
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(e.QueryStringParameters[k])
			if i < l-1 {
				buf.WriteString("&")
			}
		}
		u.RawQuery = buf.String()
	}
	return u
}

func mapHeader(e *events.APIGatewayProxyRequest) *http.Header {
	result := &http.Header{}
	for k, v := range e.Headers {
		result.Add(k, v)
	}
	return result
}

// implements https://godoc.org/net/http#ResponseWriter
type responseWriter struct {
	statusCode int
	header     http.Header
	body       *bytes.Buffer

	wroteHeader bool
	snapHeader  http.Header // snapshot of HeaderMap at first Write
}

func (rw *responseWriter) response() (*events.APIGatewayProxyResponse, error) {
	response := &events.APIGatewayProxyResponse{}

	if rw.snapHeader != nil && len(rw.snapHeader) > 0 {
		response.Headers = map[string]string{}
		for key, value := range rw.snapHeader {
			response.Headers[key] = strings.Join(value, ", ")
		}
	}

	if rw.body.Len() > 0 {
		b, err := ioutil.ReadAll(rw.body)
		if err != nil {
			return nil, err
		}
		response.Body = string(b)
	}

	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}

	response.StatusCode = rw.statusCode

	return response, nil
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) Write(buf []byte) (int, error) {
	rw.writeHeader(buf)
	if rw.body != nil {
		rw.body.Write(buf)
	}
	return len(buf), nil
}

// writes header if it hasn't been called already and trys to detect the content-type if it is not set explicitly
func (rw *responseWriter) writeHeader(b []byte) {
	if rw.wroteHeader {
		return
	}

	m := rw.Header()
	_, hasType := m["Content-Type"]
	hasTE := m.Get("Transfer-Encoding") != ""
	if !hasType && !hasTE {
		m.Set("Content-Type", http.DetectContentType(b))
	}

	rw.WriteHeader(http.StatusOK)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = statusCode
	rw.wroteHeader = true
	rw.snapHeader = cloneHeader(rw.header)
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, v := range h {
		v2 := make([]string, len(v))
		copy(v2, v)
		h2[k] = v2
	}
	return h2
}
