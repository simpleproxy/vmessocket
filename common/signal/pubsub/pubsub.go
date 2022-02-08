package pubsub

import (
	"errors"
	"sync"
	"time"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/signal/done"
	"github.com/vmessocket/vmessocket/common/task"
)

type Service struct {
	sync.RWMutex
	subs  map[string][]*Subscriber
	ctask *task.Periodic
}

type Subscriber struct {
	buffer chan interface{}
	done   *done.Instance
}

func NewService() *Service {
	s := &Service{
		subs: make(map[string][]*Subscriber),
	}
	s.ctask = &task.Periodic{
		Execute:  s.Cleanup,
		Interval: time.Second * 30,
	}
	return s
}

func (s *Service) Cleanup() error {
	s.Lock()
	defer s.Unlock()
	if len(s.subs) == 0 {
		return errors.New("nothing to do")
	}
	for name, subs := range s.subs {
		newSub := make([]*Subscriber, 0, len(s.subs))
		for _, sub := range subs {
			if !sub.IsClosed() {
				newSub = append(newSub, sub)
			}
		}
		if len(newSub) == 0 {
			delete(s.subs, name)
		} else {
			s.subs[name] = newSub
		}
	}
	if len(s.subs) == 0 {
		s.subs = make(map[string][]*Subscriber)
	}
	return nil
}

func (s *Subscriber) Close() error {
	return s.done.Close()
}

func (s *Subscriber) IsClosed() bool {
	return s.done.Done()
}

func (s *Service) Publish(name string, message interface{}) {
	s.RLock()
	defer s.RUnlock()

	for _, sub := range s.subs[name] {
		if !sub.IsClosed() {
			sub.push(message)
		}
	}
}

func (s *Subscriber) push(msg interface{}) {
	select {
	case s.buffer <- msg:
	default:
	}
}

func (s *Service) Subscribe(name string) *Subscriber {
	sub := &Subscriber{
		buffer: make(chan interface{}, 16),
		done:   done.New(),
	}
	s.Lock()
	s.subs[name] = append(s.subs[name], sub)
	s.Unlock()
	common.Must(s.ctask.Start())
	return sub
}

func (s *Subscriber) Wait() <-chan interface{} {
	return s.buffer
}
