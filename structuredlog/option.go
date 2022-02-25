package structuredlog

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type OptionBuilder struct {
	options []Option
}

type Option func(e *Event)

// newHttpFromRequest creates a http attribute from http request.
func newHttpFromRequest(req *http.Request, sc *int) *Http {
	url := req.URL.String()
	if req.URL.User != nil {
		url = strings.Replace(url, req.URL.User.String() + "@", "", -1)
	}

	var h = &Http{
		Method:    req.Method,
		URL:       url,
		Target:    req.URL.RequestURI(),
		Host:      req.URL.Hostname(),
		Scheme:    req.URL.Scheme,
		Route:     req.URL.Path,
		UserAgent: req.UserAgent(),
		ClientIP:  req.RemoteAddr,
	}

	if sc != nil {
		h.StatusCode = uint16(*sc)
	}

	return h
}

// With adds a custom option to the log event.
func (ob *OptionBuilder) With(o Option) *OptionBuilder {
	ob.options = append(ob.options, o)
	return ob
}

// WithVisibility sets the visibility to the log event.
func (ob *OptionBuilder) WithVisibility(vis bool) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if !vis {
			var visInt = 0
			e.Visibility = &visInt
		}
	})
	return ob
}

// WithName adds the name to the log event.
func (ob *OptionBuilder) WithName(name string) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		e.Name = name
	})
	return ob
}

// WithHttp adds the http attribute to the log event.
func (ob *OptionBuilder) WithHttp(http Http) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Http = &http
	})
	return ob
}

// WithHttpRequest adds the http attribute from a http request to the log event.
func (ob *OptionBuilder) WithHttpRequest(req *http.Request) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Http = newHttpFromRequest(req, nil)
	})
	return ob
}

// WithHttpResponse adds the http attribute from a http response to the log event.
func (ob *OptionBuilder) WithHttpResponse(resp *http.Response) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Http = newHttpFromRequest(resp.Request, &resp.StatusCode)
	})
	return ob
}

// WithDB adds the database attribute to the log event.
func (ob *OptionBuilder) WithDB(db DB) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.DB = &db
	})
	return ob
}

// WithException adds the exception attribute to the log event.
func (ob *OptionBuilder) WithException(err Exception) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Exception = &err
	})
	return ob
}

// With adds a custom option to the log event.
func With(o Option) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.With(o)
	return ob
}

// WithVisibility sets the visibility to the log event.
func WithVisibility(vis bool) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithVisibility(vis)
	return ob
}

// WithName adds the name to the log event.
func WithName(name string) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithName(name)
	return ob
}

// WithHttp adds the http attribute to the log event.
func WithHttp(http Http) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithHttp(http)
	return ob
}

// WithHttpRequest adds the http attribute from a http request to the log event.
func WithHttpRequest(req *http.Request) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithHttpRequest(req)
	return ob
}

// WithHttpResponse adds the http attribute from a http response to the log event.
func WithHttpResponse(resp *http.Response) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithHttpResponse(resp)
	return ob
}

// WithDB adds the database attribute to the log event.
func WithDB(db DB) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithDB(db)
	return ob
}

// WithException adds the exception attribute to the log event.
func WithException(err Exception) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithException(err)
	return ob
}

// Debug is equivalent to log.StdDebug.Print()
func (ob *OptionBuilder) Debug(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprint(v...), ob.options)
}

// Debugf is equivalent to log.StdDebug.Printf()
func (ob *OptionBuilder) Debugf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprintf(format, v...), ob.options)
}

// Info is equivalent to log.StdInfo.Print()
func (ob *OptionBuilder) Info(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprint(v...), ob.options)
}

// Infof is equivalent to log.StdInfo.Printf()
func (ob *OptionBuilder) Infof(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprintf(format, v...), ob.options)
}

// Error is equivalent to log.StdError.Print()
func (ob *OptionBuilder) Error(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprint(v...), ob.options)
}

// Errorf is equivalent to log.StdError.Printf()
func (ob *OptionBuilder) Errorf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprintf(format, v...), ob.options)
}
