package session

import (
	"encoding/json"
	"errors"
	"github.com/atsegelnyk/galaxia/model"
)

const DefaultSessionTTL = 86400

type Session struct {
	TTL              int64                        `json:"ttl"`
	UserID           int64                        `json:"user_id"`
	CurrentStage     model.ResourceRef            `json:"current_stage"`
	PendingCallbacks map[string]model.ResourceRef `json:"pending_callbacks"`

	Misc map[string]string `json:"misc"`
}

func NewSession(userID int64) *Session {
	return &Session{
		UserID:           userID,
		TTL:              DefaultSessionTTL,
		PendingCallbacks: make(map[string]model.ResourceRef),
	}
}

func (s *Session) GetUserID() int64 {
	return s.UserID
}

func (s *Session) GetCurrentStage() model.ResourceRef {
	return s.CurrentStage
}

func (s *Session) SetStage(nextStageRef model.ResourceRef) {
	s.CurrentStage = nextStageRef
}

func (s *Session) RegisterCallback(id string, handlerRef model.ResourceRef) {
	s.PendingCallbacks[id] = handlerRef
}

func (s *Session) GetPendingCallback(callbackID string) (model.ResourceRef, error) {
	if cb, ok := s.PendingCallbacks[callbackID]; ok {
		return cb, nil
	}
	return "", errors.New("callback not found")
}

func (s *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Session) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
