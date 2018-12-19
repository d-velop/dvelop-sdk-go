package lambda

import (
	"context"
	"errors"
)

type contextKey string

const reqIdCtxKey = contextKey("reqId")

// AddReqIdToCtx adds the lambda request ID to the context
func AddReqIdToCtx(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, reqIdCtxKey, reqId)
}

// ReqIdFromCtx reads the lambda request ID from the context
func ReqIdFromCtx(ctx context.Context) (string, error) {
	reqId, ok := ctx.Value(reqIdCtxKey).(string)
	if !ok {
		return "", errors.New("no requestid on context")
	}
	return reqId, nil
}
