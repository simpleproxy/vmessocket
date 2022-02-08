package protocol

import (
	"sync"
	"time"

	"github.com/vmessocket/vmessocket/common/dice"
	"github.com/vmessocket/vmessocket/common/net"
)

type alwaysValidStrategy struct{}

type ServerSpec struct {
	sync.RWMutex
	dest  net.Destination
	users []*MemoryUser
	valid ValidationStrategy
}

type timeoutValidStrategy struct {
	until time.Time
}

type ValidationStrategy interface {
	IsValid() bool
	Invalidate()
}

func AlwaysValid() ValidationStrategy {
	return alwaysValidStrategy{}
}

func BeforeTime(t time.Time) ValidationStrategy {
	return &timeoutValidStrategy{
		until: t,
	}
}

func NewServerSpec(dest net.Destination, valid ValidationStrategy, users ...*MemoryUser) *ServerSpec {
	return &ServerSpec{
		dest:  dest,
		users: users,
		valid: valid,
	}
}

func NewServerSpecFromPB(spec *ServerEndpoint) (*ServerSpec, error) {
	dest := net.TCPDestination(spec.Address.AsAddress(), net.Port(spec.Port))
	mUsers := make([]*MemoryUser, len(spec.User))
	for idx, u := range spec.User {
		mUser, err := u.ToMemoryUser()
		if err != nil {
			return nil, err
		}
		mUsers[idx] = mUser
	}
	return NewServerSpec(dest, AlwaysValid(), mUsers...), nil
}

func (s *ServerSpec) AddUser(user *MemoryUser) {
	if s.HasUser(user) {
		return
	}
	s.Lock()
	defer s.Unlock()
	s.users = append(s.users, user)
}

func (s *ServerSpec) Destination() net.Destination {
	return s.dest
}

func (s *ServerSpec) HasUser(user *MemoryUser) bool {
	s.RLock()
	defer s.RUnlock()

	for _, u := range s.users {
		if u.Account.Equals(user.Account) {
			return true
		}
	}
	return false
}

func (alwaysValidStrategy) Invalidate() {}

func (alwaysValidStrategy) IsValid() bool {
	return true
}

func (s *ServerSpec) IsValid() bool {
	return s.valid.IsValid()
}

func (s *timeoutValidStrategy) IsValid() bool {
	return s.until.After(time.Now())
}

func (s *ServerSpec) Invalidate() {
	s.valid.Invalidate()
}

func (s *timeoutValidStrategy) Invalidate() {
	s.until = time.Time{}
}

func (s *ServerSpec) PickUser() *MemoryUser {
	s.RLock()
	defer s.RUnlock()
	userCount := len(s.users)
	switch userCount {
	case 0:
		return nil
	case 1:
		return s.users[0]
	default:
		return s.users[dice.Roll(userCount)]
	}
}
