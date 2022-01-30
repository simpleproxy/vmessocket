package common

import "github.com/vmessocket/vmessocket/common/errors"

type Closable interface {
	Close() error
}

type Interruptible interface {
	Interrupt()
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

type Runnable interface {
	Start() error
	Closable
}

type HasType interface {
	Type() interface{}
}

type ChainedClosable []Closable

func (cc ChainedClosable) Close() error {
	var errs []error
	for _, c := range cc {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Combine(errs...)
}
