// Package tracecontext contains functions to handle a trace context for the current request.
//
// A trace context is based on the W3C trace-context specification available at https://w3c.github.io/trace-context/.
//
// Logstatements should log the trace-id and span-id in every statement to correlate
// the statements to a specific request. This simplifies the tracking of
// a request through a system which serves multiple concurrent requests.
package tracecontext

import (
	"context"
	"errors"
	"net/http"
)

type contextKey string

const traceIdCtxKey = contextKey("traceId")
const spanIdCtxKey = contextKey("spanId")
const traceparentHeader = "traceparent"

// AddToCtx reads the http header traceparent from the current request
// and stores the trace-id in the context. The span-id is regenerated on request.
// If the request doesn't have an existing trace-id a new one is generated.
func AddToCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			if traceId, err := getTraceId(req.Header); err == nil {
				ctx = WithTraceIdCtx(ctx, traceId)
			}

			if spanId, err := NewSpanId(); err == nil {
				ctx = WithSpanIdCtx(ctx, spanId)
			}

			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

// TraceparentFromCtx reads the current trace-id and span-id from the context and builds the traceparent header value.
func TraceparentFromCtx(ctx context.Context) (string, error) {
	traceId, err := TraceIdFromCtx(ctx)
	if err != nil {
		return "", err
	}
	spanId, err := SpanIdFromCtx(ctx)
	if err != nil {
		return "", err
	}

	tp, err := NewTraceparent(traceId, spanId)
	if err != nil {
		return "", err
	}

	return tp.String(), nil
}

// TraceIdFromCtx reads the current trace-id from the context.
func TraceIdFromCtx(ctx context.Context) (string, error) {
	traceId, ok := ctx.Value(traceIdCtxKey).(string)
	if !ok {
		return "", errors.New("no traceId on context")
	}
	return traceId, nil
}

// SpanIdFromCtx reads the current span-id from the context.
func SpanIdFromCtx(ctx context.Context) (string, error) {
	spanId, ok := ctx.Value(spanIdCtxKey).(string)
	if !ok {
		return "", errors.New("no spanId on context")
	}
	return spanId, nil
}

// WithTraceIdCtx returns a new context.Context with the given trace id
func WithTraceIdCtx(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, traceIdCtxKey, traceId)
}

// WithSpanIdCtx returns a new context.Context with the given span id
func WithSpanIdCtx(ctx context.Context, spanId string) context.Context {
	return context.WithValue(ctx, spanIdCtxKey, spanId)
}

func getTraceId(header http.Header) (string, error) {
	if s := header.Get(traceparentHeader); s != "" {
		if t, err := ParseTraceparent(s); err == nil {
			return t.TraceId(), nil
		}
	}

	traceId, err := NewTraceId()
	if err != nil {
		return "", err
	}
	return traceId, nil
}