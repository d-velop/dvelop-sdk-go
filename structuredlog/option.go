package structuredlog

import (
	"context"
	"fmt"
)

type OptionBuilder struct {
	options []Option
}

type Option func(e *Event)

// With adds a custom option of the log event.
func (ob *OptionBuilder) With(o Option) *OptionBuilder {
	ob.options = append(ob.options, o)
	return ob
}

// WithVisibility sets the visibility of the log event.
func (ob *OptionBuilder) WithVisibility(vis bool) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if !vis {
			var visInt = 0
			e.Visibility = &visInt
		}
	})
	return ob
}

// WithName adds the name of the log event.
func (ob *OptionBuilder) WithName(name string) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		e.Name = name
	})
	return ob
}

// WithHttp adds the http attribute of the log event.
func (ob *OptionBuilder) WithHttp(http Http) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Http = &http
	})
	return ob
}

// WithDB adds the database attribute of the log event.
func (ob *OptionBuilder) WithDB(db DB) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.DB = &db
	})
	return ob
}

// WithException adds the exception attribute of the log event.
func (ob *OptionBuilder) WithException(err Exception) *OptionBuilder {
	ob.options = append(ob.options, func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Exception = &err
	})
	return ob
}

// With adds a custom option of the log event.
func With(o Option) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.With(o)
	return ob
}

// WithVisibility sets the visibility of the log event.
func WithVisibility(vis bool) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithVisibility(vis)
	return ob
}

// WithName adds the name of the log event.
func WithName(name string) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithName(name)
	return ob
}

// WithHttp adds the http attribute of the log event.
func WithHttp(http Http) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithHttp(http)
	return ob
}

// WithDB adds the database attribute of the log event.
func WithDB(db DB) *OptionBuilder {
	ob := &OptionBuilder{}
	ob.WithDB(db)
	return ob
}

// WithException adds the exception attribute of the log event.
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
