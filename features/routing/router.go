package routing

import "github.com/vmessocket/vmessocket/features"

type DefaultRouter struct{}

type Router interface {
	features.Feature
}

func RouterType() interface{} {
	return (*Router)(nil)
}

func (DefaultRouter) Close() error {
	return nil
}

func (DefaultRouter) Start() error {
	return nil
}

func (DefaultRouter) Type() interface{} {
	return RouterType()
}
