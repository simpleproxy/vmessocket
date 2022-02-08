package core

import "context"

const vmessocketKey vmessocketKeyType = 1

type temporaryValueDelegationFix struct {
	context.Context
	value context.Context
}

type vmessocketKeyType int

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

func ToBackgroundDetachedContext(ctx context.Context) context.Context {
	return &temporaryValueDelegationFix{context.Background(), ctx}
}

func toContext(ctx context.Context, v *Instance) context.Context {
	if FromContext(ctx) != v {
		ctx = context.WithValue(ctx, vmessocketKey, v)
	}
	return ctx
}

func (t *temporaryValueDelegationFix) Value(key interface{}) interface{} {
	return t.value.Value(key)
}
