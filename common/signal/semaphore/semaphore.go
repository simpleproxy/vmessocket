package semaphore

type Instance struct {
	token chan struct{}
}

func New(n int) *Instance {
	s := &Instance{
		token: make(chan struct{}, n),
	}
	for i := 0; i < n; i++ {
		s.token <- struct{}{}
	}
	return s
}

func (s *Instance) Wait() <-chan struct{} {
	return s.token
}

func (s *Instance) Signal() {
	s.token <- struct{}{}
}
