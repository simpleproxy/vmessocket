package done

import (
	"sync"
)

type Instance struct {
	access sync.Mutex
	c      chan struct{}
	closed bool
}

func New() *Instance {
	return &Instance{
		c: make(chan struct{}),
	}
}

func (d *Instance) Done() bool {
	select {
	case <-d.Wait():
		return true
	default:
		return false
	}
}

func (d *Instance) Wait() <-chan struct{} {
	return d.c
}

func (d *Instance) Close() error {
	d.access.Lock()
	defer d.access.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	close(d.c)

	return nil
}
