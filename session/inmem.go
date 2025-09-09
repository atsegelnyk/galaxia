package session

import (
	"sync"
	"time"
)

type InMemorySessionRepository struct {
	mu        sync.Mutex
	sessions  map[int64]*Session
	expireMap map[int64][]int64
}

func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		mu:        sync.Mutex{},
		sessions:  make(map[int64]*Session),
		expireMap: make(map[int64][]int64),
	}
}

func (m *InMemorySessionRepository) Get(userID int64) (*Session, error) {
	if session, ok := m.sessions[userID]; ok {
		return session, nil
	}
	return nil, NotFoundError
}

func (m *InMemorySessionRepository) Save(session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.UserID] = session
	m.expireMap[session.ExpireTime.Unix()] = append(m.expireMap[session.ExpireTime.Unix()], session.UserID)
	return nil
}

func (m *InMemorySessionRepository) Expire(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, userID)
}

func (m *InMemorySessionRepository) expirationWorker() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		ts := time.Now().Unix()
		if userIDs, ok := m.expireMap[ts]; ok {
			m.mu.Lock()
			for _, userID := range userIDs {
				delete(m.sessions, userID)
			}
			delete(m.expireMap, ts)
			m.mu.Unlock()
		}
	}
}
