package protocol

import (
	"context"
)

const requestKey key = iota

type key int

func ContextWithRequestHeader(ctx context.Context, request *RequestHeader) context.Context {
	return context.WithValue(ctx, requestKey, request)
}

func RequestHeaderFromContext(ctx context.Context) *RequestHeader {
	request := ctx.Value(requestKey)
	if request == nil {
		return nil
	}
	return request.(*RequestHeader)
}
