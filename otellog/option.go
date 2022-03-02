package otellog

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type LogBuilder struct {
	options []Option
}

type Option func(e *Event)

// newHttpFromRequest creates a http attribute from http request.
func newHttpFromRequest(req *http.Request, sc *int) *Http {
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

	var h = &Http{
		Method:    req.Method,
		URL:       url,
		Target:    req.URL.RequestURI(),
		Host:      req.URL.Hostname(),
		Scheme:    req.URL.Scheme,
		Route:     req.URL.Path,
		UserAgent: req.UserAgent(),
		ClientIP:  ipAddress,
	}

	if sc != nil {
		h.StatusCode = uint16(*sc)
	}

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
		e.Attributes.Http = newHttpFromRequest(req, nil)
	})
	return ob
}

// WithHttpResponse adds the http attribute from a http response to the log event.
func (ob *LogBuilder) WithHttpResponse(resp *http.Response) *LogBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Http = newHttpFromRequest(resp.Request, &resp.StatusCode)
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

// WithException adds the exception attribute to the log event.
func WithException(err Exception) *LogBuilder {
	ob := &LogBuilder{}
	ob.WithException(err)
	return ob
}

// Debug is equivalent to log.StdDebug.Print()
func (ob *LogBuilder) Debug(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprint(v...), ob.options)
}

// Debugf is equivalent to log.StdDebug.Printf()
func (ob *LogBuilder) Debugf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprintf(format, v...), ob.options)
}

// Info is equivalent to log.StdInfo.Print()
func (ob *LogBuilder) Info(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprint(v...), ob.options)
}

// Infof is equivalent to log.StdInfo.Printf()
func (ob *LogBuilder) Infof(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprintf(format, v...), ob.options)
}

// Error is equivalent to log.StdError.Print()
func (ob *LogBuilder) Error(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprint(v...), ob.options)
}

// Errorf is equivalent to log.StdError.Printf()
func (ob *LogBuilder) Errorf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprintf(format, v...), ob.options)
}
