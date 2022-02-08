package routing

import (
	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/features"
)

type DefaultRouter struct{}

type Route interface {
	Context
	GetOutboundGroupTags() []string
	GetOutboundTag() string
}

type Router interface {
	features.Feature
	PickRoute(ctx Context) (Route, error)
}

func RouterType() interface{} {
	return (*Router)(nil)
}

func (DefaultRouter) Close() error {
	return nil
}

func (DefaultRouter) PickRoute(ctx Context) (Route, error) {
	return nil, common.ErrNoClue
}

func (DefaultRouter) Start() error {
	return nil
}

func (DefaultRouter) Type() interface{} {
	return RouterType()
}
