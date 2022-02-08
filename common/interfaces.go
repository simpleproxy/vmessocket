package common

import "github.com/vmessocket/vmessocket/common/errors"

type ChainedClosable []Closable

type Closable interface {
	Close() error
}

type HasType interface {
	Type() interface{}
}

type Interruptible interface {
	Interrupt()
}

type Runnable interface {
	Start() error
	Closable
}

func Close(obj interface{}) error {
	if c, ok := obj.(Closable); ok {
		return c.Close()
	}
	return nil
}

func Interrupt(obj interface{}) error {
	if c, ok := obj.(Interruptible); ok {
		c.Interrupt()
		return nil
	}
	return Close(obj)
}

func (cc ChainedClosable) Close() error {
	var errs []error
	for _, c := range cc {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Combine(errs...)
}
