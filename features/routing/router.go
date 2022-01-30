package routing

import (
	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/features"
)

type Router interface {
	features.Feature
	PickRoute(ctx Context) (Route, error)
}

type Route interface {
	Context
	GetOutboundGroupTags() []string
	GetOutboundTag() string
}

func RouterType() interface{} {
	return (*Router)(nil)
}

type DefaultRouter struct{}

func (DefaultRouter) Type() interface{} {
	return RouterType()
}

func (DefaultRouter) PickRoute(ctx Context) (Route, error) {
	return nil, common.ErrNoClue
}

func (DefaultRouter) Start() error {
	return nil
}

func (DefaultRouter) Close() error {
	return nil
}
