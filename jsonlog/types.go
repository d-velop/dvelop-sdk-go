package jsonlog

import (
	"context"
	"time"
)

type Time func() time.Time

type Hook func(ctx context.Context, event *Event)

type Writer func(event *Event) ([]byte, error)

type LogOption func(e *Event)

func WithName(name string) LogOption {
	return func(e *Event) {
		e.Name = name
	}
}

func WithVisibility(vis bool) LogOption {
	return func(e *Event) {
		if !vis {
			var visInt = 0
			e.Visibility = &visInt
		}
	}
}

func WithHttp(http Http) LogOption {
	return func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Http = &http
	}
}

func WithDB(db DB) LogOption {
	return func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.DB = &db
	}
}

func WithException(err Exception) LogOption {
	return func(e *Event) {
		if e.Attributes == nil {
			e.Attributes = &Attributes{}
		}
		e.Attributes.Exception = &err
	}
}
