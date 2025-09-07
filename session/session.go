package session

import (
	"github.com/atsegelnyk/galaxia/model"
	"sync"
)

type Session struct {
	userID   int64
	stageRef *model.Stage
}

func NewSession(userID int64) *Session {
	return &Session{
		userID: userID,
	}
}

func (s *Session) GetUserID() int64 {
	return s.userID
}

func (s *Session) GetCurrentStage() *model.Stage {
	return s.stageRef
}

func (s *Session) SetStage(stageRef *model.Stage) {
	s.stageRef = stageRef
}

type Manager struct {
	mu       sync.Mutex
	sessions map[int64]*Session
}

func NewManager() *Manager {
	return &Manager{
		mu:       sync.Mutex{},
		sessions: make(map[int64]*Session),
	}
}

func (m *Manager) GetForUserID(userID int64) *Session {
	if session, ok := m.sessions[userID]; ok {
		return session
	}
	session := NewSession(userID)
	m.mu.Lock()
	m.sessions[userID] = session
	m.mu.Unlock()
	return session
}
