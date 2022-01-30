package stats

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

import (
	"context"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/features"
)

type Counter interface {
	Value() int64
	Set(int64) int64
	Add(int64) int64
}

type Channel interface {
	common.Runnable
	Publish(context.Context, interface{})
	Subscribers() []chan interface{}
	Subscribe() (chan interface{}, error)
	Unsubscribe(chan interface{}) error
}

func SubscribeRunnableChannel(c Channel) (chan interface{}, error) {
	if len(c.Subscribers()) == 0 {
		if err := c.Start(); err != nil {
			return nil, err
		}
	}
	return c.Subscribe()
}

func UnsubscribeClosableChannel(c Channel, sub chan interface{}) error {
	if err := c.Unsubscribe(sub); err != nil {
		return err
	}
	if len(c.Subscribers()) == 0 {
		return c.Close()
	}
	return nil
}

type Manager interface {
	features.Feature
	RegisterCounter(string) (Counter, error)
	UnregisterCounter(string) error
	GetCounter(string) Counter
	RegisterChannel(string) (Channel, error)
	UnregisterChannel(string) error
	GetChannel(string) Channel
}

func GetOrRegisterCounter(m Manager, name string) (Counter, error) {
	counter := m.GetCounter(name)
	if counter != nil {
		return counter, nil
	}

	return m.RegisterCounter(name)
}

func GetOrRegisterChannel(m Manager, name string) (Channel, error) {
	channel := m.GetChannel(name)
	if channel != nil {
		return channel, nil
	}

	return m.RegisterChannel(name)
}

func ManagerType() interface{} {
	return (*Manager)(nil)
}

type NoopManager struct{}

func (NoopManager) Type() interface{} {
	return ManagerType()
}

func (NoopManager) RegisterCounter(string) (Counter, error) {
	return nil, newError("not implemented")
}

func (NoopManager) UnregisterCounter(string) error {
	return nil
}

func (NoopManager) GetCounter(string) Counter {
	return nil
}

func (NoopManager) RegisterChannel(string) (Channel, error) {
	return nil, newError("not implemented")
}

func (NoopManager) UnregisterChannel(string) error {
	return nil
}

func (NoopManager) GetChannel(string) Channel {
	return nil
}

func (NoopManager) Start() error { return nil }

func (NoopManager) Close() error { return nil }
