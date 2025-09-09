package session

import (
	"encoding/json"
	"errors"
	"github.com/atsegelnyk/galaxia/model"
	"time"
)

const DefaultSessionTTL = 86400

type Session struct {
	ExpireTime       time.Time
	TTL              int64                             `json:"ttl"`
	UserID           int64                             `json:"user_id"`
	CurrentStage     model.ResourceRef                 `json:"current_stage"`
	UserContext      *model.UserContext                `json:"context"`
	PendingCallbacks map[string]*model.PendingCallback `json:"pending_callbacks"`
	StageMessages    []int                             `json:"pending_messages"`
}

func NewSession(userID int64, opts ...Option) *Session {
	baseSession := &Session{
		UserID: userID,
		UserContext: &model.UserContext{
			UserID: userID,
		},
		TTL:              DefaultSessionTTL,
		ExpireTime:       time.Now().Add(time.Duration(DefaultSessionTTL) * time.Second),
		PendingCallbacks: make(map[string]*model.PendingCallback),
	}
	for _, opt := range opts {
		opt(baseSession)
	}
	return baseSession
}

type Option func(*Session)

func WithLang(lang string) Option {
	return func(session *Session) {
		session.UserContext.Lang = lang
	}
}

func WithUsername(username string) Option {
	return func(session *Session) {
		session.UserContext.Username = username
	}
}

func WithName(name string) Option {
	return func(session *Session) {
		session.UserContext.Name = name
	}
}

func WithLastName(lastName string) Option {
	return func(session *Session) {
		session.UserContext.LastName = lastName
	}
}

func WithTTL(ttl int64) Option {
	return func(session *Session) {
		session.TTL = ttl
		session.ExpireTime = time.Now().Add(time.Duration(ttl) * time.Second)
	}
}

func (s *Session) GetCurrentStage() model.ResourceRef {
	return s.CurrentStage
}

func (s *Session) SetNextStage(nextStageRef model.ResourceRef) {
	s.CurrentStage = nextStageRef
}

func (s *Session) RegisterCallback(id string, behaviour model.CallbackBehaviour, handlerRef model.ResourceRef) {
	s.PendingCallbacks[id] = &model.PendingCallback{
		Behaviour:  behaviour,
		HandlerRef: handlerRef,
	}
}

func (s *Session) GetPendingCallbackHandler(callbackID string) (model.ResourceRef, error) {
	if cb, ok := s.PendingCallbacks[callbackID]; ok {
		ref := cb.HandlerRef
		if cb.Behaviour == model.DeleteCallbackBehaviour {
			delete(s.PendingCallbacks, callbackID)
		}
		return ref, nil
	}
	return "", errors.New("callback not found")
}

func (s *Session) AppendStageMessage(msgID int) {
	s.StageMessages = append(s.StageMessages, msgID)
}

func (s *Session) Clean() {
	s.StageMessages = nil
	s.PendingCallbacks = make(map[string]*model.PendingCallback)
}

func (s *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Session) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
