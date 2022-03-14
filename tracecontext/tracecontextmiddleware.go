package tracecontext

import (
	"context"
	"errors"
	"net/http"
)

type contextKey string

const traceIdCtxKey = contextKey("traceId")
const spanIdCtxKey = contextKey("spanId")
const traceParentHeader = "traceparent"

func AddToCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = withTraceIdCtx(ctx, req.Header)
			ctx = withSpanIdCtx(ctx)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

func TraceParentFromCtx(ctx context.Context) (string, error) {
	traceId, err := TraceIdFromCtx(ctx)
	if err != nil {
		return "", err
	}
	spanId, err := SpanIdFromCtx(ctx)
	if err != nil {
		return "", err
	}

	tp, err := NewTraceParent(traceId, spanId)
	if err != nil {
		return "", err
	}

	return tp.String(), nil
}

func TraceIdFromCtx(ctx context.Context) (string, error) {
	traceId, ok := ctx.Value(traceIdCtxKey).(string)
	if !ok {
		return "", errors.New("no traceId on context")
	}
	return traceId, nil
}

func SpanIdFromCtx(ctx context.Context) (string, error) {
	spanId, ok := ctx.Value(spanIdCtxKey).(string)
	if !ok {
		return "", errors.New("no spanId on context")
	}
	return spanId, nil
}

func withTraceIdCtx(parent context.Context, header http.Header) context.Context {
	if s := header.Get(traceParentHeader); s != "" {
		if t, err := ParseTraceParent(s); err == nil {
			return context.WithValue(parent, traceIdCtxKey, t.TraceId())
		}
	}
	if traceId, err := NewTraceId(); err == nil {
		return context.WithValue(parent, traceIdCtxKey, traceId)
	}
	return parent
}

func withSpanIdCtx(parent context.Context) context.Context {
	if spanId, err := NewSpanId(); err == nil {
		return context.WithValue(parent, spanIdCtxKey, spanId)
	}
	return parent
}
