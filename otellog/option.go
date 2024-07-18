package otellog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type LogBuilder struct {
	options []Option
}

type Option func(e *Event)

// enrichHttpAttributeWithRequest enriches http attribute with http request.
func enrichHttpAttributeWithRequest(h *Http, req *http.Request) *Http {
	url := req.URL.String()
	if req.URL.User != nil {
		url = strings.Replace(url, req.URL.User.String()+"@", "", -1)
	}

	ipAddress := req.RemoteAddr
	fwdAddress := req.Header.Get("X-Forwarded-For")
	if fwdAddress != "" {
		ipAddress = fwdAddress
		ips := strings.Split(fwdAddress, ", ")
		if len(ips) > 1 {
			ipAddress = ips[0]
		}
	}

	h.Method = req.Method
	h.URL = url
	h.Target = req.URL.RequestURI()
	h.Host = req.URL.Hostname()
	h.Scheme = req.URL.Scheme
	h.Route = req.URL.Path
	h.UserAgent = req.UserAgent()
	h.ClientIP = ipAddress
	return h
}

// With adds a custom option to the log event.
func (ob *LogBuilder) With(o Option) *LogBuilder {
	ob.options = append(ob.options, o)
	return ob
}

// WithVisibility sets the visibility to the log event.
func (ob *LogBuilder) WithVisibility(vis bool) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if !vis {
			var visInt = 0
			e.Visibility = &visInt
		}
	})
	return ob
}

// WithName adds the name to the log event.
func (ob *LogBuilder) WithName(name string) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		e.Name = name
	})
	return ob
}

// WithHttp adds the http attribute to the log event.
func (ob *LogBuilder) WithHttp(http Http) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Http = &http
	})
	return ob
}

// WithHttpRequest adds the http attribute from a http request to the log event.
func (ob *LogBuilder) WithHttpRequest(req *http.Request) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		if e.Attributes.Http == nil {
			e.Attributes.Http = &Http{}
		}
		enrichHttpAttributeWithRequest(e.Attributes.Http, req)
	})
	return ob
}

// WithHttpResponse adds the http attribute from a http response to the log event.
func (ob *LogBuilder) WithHttpResponse(resp *http.Response) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		if e.Attributes.Http == nil {
			e.Attributes.Http = &Http{}
		}
		enrichHttpAttributeWithRequest(e.Attributes.Http, resp.Request)
		e.Attributes.Http.StatusCode = uint16(resp.StatusCode)
	})
	return ob
}

// WithHttpStatusCode adds the http status code to the log event.
func (ob *LogBuilder) WithHttpStatusCode(status int) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		if e.Attributes.Http == nil {
			e.Attributes.Http = &Http{}
		}
		e.Attributes.Http.StatusCode = uint16(status)
	})
	return ob
}

// WithDB adds the database attribute to the log event.
func (ob *LogBuilder) WithDB(db DB) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.DB = &db
	})
	return ob
}

func (ob *LogBuilder) WithUserId(userId string) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if userId != "" {
			hash := sha256.New()
			hash.Write([]byte(userId))
			sum := hash.Sum(nil)
			uidHash := hex.EncodeToString(sum)
			e.UserIdHash = uidHash
		}
	})
	return ob
}

// WithException adds the exception attribute to the log event.
func (ob *LogBuilder) WithException(err Exception) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Exception = &err
	})
	return ob
}

// WithAdditionalAttributes adds custom attributes to the log event.
func (ob *LogBuilder) WithAdditionalAttributes(additionalAttr interface{}) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		err := e.Attributes.AddAdditionalAttributes(additionalAttr)
		if err != nil {
			panic(err)
		}
	})
	return ob
}

// With adds a custom option to the log event.
func With(o Option) *LogBuilder {
	ob := &LogBuilder{}
	ob.With(o)
	return ob
}

// WithVisibility sets the visibility to the log event.
func WithVisibility(vis bool) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithVisibility(vis)
	return ob
}

// WithName adds the name to the log event.
func WithName(name string) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithName(name)
	return ob
}

// WithHttp adds the http attribute to the log event.
func WithHttp(http Http) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithHttp(http)
	return ob
}

// WithHttpRequest adds the http attribute from a http request to the log event.
func WithHttpRequest(req *http.Request) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithHttpRequest(req)
	return ob
}

// WithHttpStatusCode adds the http status code to the log event.
func WithHttpStatusCode(status int) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithHttpStatusCode(status)
	return ob
}

// WithHttpResponse adds the http attribute from a http response to the log event.
func WithHttpResponse(resp *http.Response) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithHttpResponse(resp)
	return ob
}

// WithDB adds the database attribute to the log event.
func WithDB(db DB) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithDB(db)
	return ob
}

// WithUserId hashes the userId and adds the uid attribute to the log event.
func WithUserId(userId string) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithUserId(userId)
	return ob
}

// WithException adds the exception attribute to the log event.
func WithException(err Exception) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithException(err)
	return ob
}

// WithAdditionalAttributes adds the exception attribute to the log event.
func WithAdditionalAttributes(additionalAttr interface{}) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithAdditionalAttributes(additionalAttr)
	return ob
}

// Debug logs an event body according to the otel definition
func (ob *LogBuilder) Debug(ctx context.Context, body interface{}) {
	std.output(ctx, SeverityDebug, body, ob.options)
}

// Debugf is equivalent to log.StdDebug.Printf()
func (ob *LogBuilder) Debugf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprintf(format, v...), ob.options)
}

// Info logs an event body according to the otel definition
func (ob *LogBuilder) Info(ctx context.Context, body interface{}) {
	std.output(ctx, SeverityInfo, body, ob.options)
}

// Infof is equivalent to log.StdInfo.Printf()
func (ob *LogBuilder) Infof(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprintf(format, v...), ob.options)
}

// Error logs an event body according to the otel definition
func (ob *LogBuilder) Error(ctx context.Context, body interface{}) {
	std.output(ctx, SeverityError, body, ob.options)
}

// Errorf is equivalent to log.StdError.Printf()
func (ob *LogBuilder) Errorf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprintf(format, v...), ob.options)
}
