package mux

type Session struct {
	ID uint16
}

type SessionManager struct {
	sessions map[uint16]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[uint16]*Session, 16),
	}
}
