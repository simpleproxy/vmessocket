//go:build !confonly
// +build !confonly

package core

import (
	"context"
)

type vmessocketKeyType int

const vmessocketKey vmessocketKeyType = 1

func FromContext(ctx context.Context) *Instance {
	if s, ok := ctx.Value(vmessocketKey).(*Instance); ok {
		return s
	}
	return nil
}

func MustFromContext(ctx context.Context) *Instance {
	v := FromContext(ctx)
	if v == nil {
		panic("V is not in context.")
	}
	return v
}

func toContext(ctx context.Context, v *Instance) context.Context {
	if FromContext(ctx) != v {
		ctx = context.WithValue(ctx, vmessocketKey, v)
	}
	return ctx
}

func ToBackgroundDetachedContext(ctx context.Context) context.Context {
	return &temporaryValueDelegationFix{context.Background(), ctx}
}

type temporaryValueDelegationFix struct {
	context.Context
	value context.Context
}

func (t *temporaryValueDelegationFix) Value(key interface{}) interface{} {
	return t.value.Value(key)
}
