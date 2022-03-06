package mux

import (
	"sync"
)

type Session struct {
	ID uint16
}

type SessionManager struct {
	sync.RWMutex
	sessions map[uint16]*Session
	count    uint16
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[uint16]*Session, 16),
	}
}
