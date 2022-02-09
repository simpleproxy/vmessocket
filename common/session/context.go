package session

import "context"

const (
	idSessionKey sessionKey = iota
	inboundSessionKey
	outboundSessionKey
	contentSessionKey
	sockoptSessionKey
	trackedConnectionErrorKey
)

type sessionKey int

type TrackedRequestErrorFeedback interface {
	SubmitError(err error)
}

func ContentFromContext(ctx context.Context) *Content {
	if content, ok := ctx.Value(contentSessionKey).(*Content); ok {
		return content
	}
	return nil
}

func ContextWithContent(ctx context.Context, content *Content) context.Context {
	return context.WithValue(ctx, contentSessionKey, content)
}

func ContextWithID(ctx context.Context, id ID) context.Context {
	return context.WithValue(ctx, idSessionKey, id)
}

func ContextWithInbound(ctx context.Context, inbound *Inbound) context.Context {
	return context.WithValue(ctx, inboundSessionKey, inbound)
}

func ContextWithOutbound(ctx context.Context, outbound *Outbound) context.Context {
	return context.WithValue(ctx, outboundSessionKey, outbound)
}

func ContextWithSockopt(ctx context.Context, s *Sockopt) context.Context {
	return context.WithValue(ctx, sockoptSessionKey, s)
}

func GetForcedOutboundTagFromContext(ctx context.Context) string {
	if ContentFromContext(ctx) == nil {
		return ""
	}
	return ContentFromContext(ctx).Attribute("forcedOutboundTag")
}

func GetTransportLayerProxyTagFromContext(ctx context.Context) string {
	if ContentFromContext(ctx) == nil {
		return ""
	}
	return ContentFromContext(ctx).Attribute("transportLayerOutgoingTag")
}

func IDFromContext(ctx context.Context) ID {
	if id, ok := ctx.Value(idSessionKey).(ID); ok {
		return id
	}
	return 0
}

func InboundFromContext(ctx context.Context) *Inbound {
	if inbound, ok := ctx.Value(inboundSessionKey).(*Inbound); ok {
		return inbound
	}
	return nil
}

func OutboundFromContext(ctx context.Context) *Outbound {
	if outbound, ok := ctx.Value(outboundSessionKey).(*Outbound); ok {
		return outbound
	}
	return nil
}

func SetForcedOutboundTagToContext(ctx context.Context, tag string) context.Context {
	if contentFromContext := ContentFromContext(ctx); contentFromContext == nil {
		ctx = ContextWithContent(ctx, &Content{})
	}
	ContentFromContext(ctx).SetAttribute("forcedOutboundTag", tag)
	return ctx
}

func SetTransportLayerProxyTagToContext(ctx context.Context, tag string) context.Context {
	if contentFromContext := ContentFromContext(ctx); contentFromContext == nil {
		ctx = ContextWithContent(ctx, &Content{})
	}
	ContentFromContext(ctx).SetAttribute("transportLayerOutgoingTag", tag)
	return ctx
}

func SockoptFromContext(ctx context.Context) *Sockopt {
	if sockopt, ok := ctx.Value(sockoptSessionKey).(*Sockopt); ok {
		return sockopt
	}
	return nil
}

func SubmitOutboundErrorToOriginator(ctx context.Context, err error) {
	if errorTracker := ctx.Value(trackedConnectionErrorKey); errorTracker != nil {
		errorTracker := errorTracker.(TrackedRequestErrorFeedback)
		errorTracker.SubmitError(err)
	}
}

func TrackedConnectionError(ctx context.Context, tracker TrackedRequestErrorFeedback) context.Context {
	return context.WithValue(ctx, trackedConnectionErrorKey, tracker)
}
