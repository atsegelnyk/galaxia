package session

import (
	"encoding/json"
	"errors"
	"github.com/atsegelnyk/galaxia/model"
	sessionpb "github.com/atsegelnyk/galaxia/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	PendingInputs    map[string]model.ResourceRef      `json:"pending_inputs"`
	StageMessages    []int64                           `json:"pending_messages"`
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
		PendingInputs:    make(map[string]model.ResourceRef),
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

func (s *Session) AppendStageMessage(msgID int64) {
	s.StageMessages = append(s.StageMessages, msgID)
}

func (s *Session) Clean() {
	s.StageMessages = nil
	s.PendingCallbacks = make(map[string]*model.PendingCallback)
}

func (s *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(*s)
}

func (s *Session) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *Session) MarshalProto() ([]byte, error) {
	var pbCtx *sessionpb.UserContext
	if s.UserContext != nil {
		var misc *structpb.Struct
		if s.UserContext.Misc != nil {
			m, err := structpb.NewStruct(s.UserContext.Misc)
			if err != nil {
				return nil, err
			}
			misc = m
		}
		pbCtx = &sessionpb.UserContext{
			UserId:   s.UserContext.UserID,
			Lang:     s.UserContext.Lang,
			Name:     s.UserContext.Name,
			LastName: s.UserContext.LastName,
			Username: s.UserContext.Username,
			Misc:     misc,
		}
	}

	pbCbs := make(map[string]*sessionpb.PendingCallback, len(s.PendingCallbacks))
	for k, v := range s.PendingCallbacks {
		if v == nil {
			continue
		}
		pbCbs[k] = &sessionpb.PendingCallback{
			HandlerRef: string(v.HandlerRef),
			Behaviour:  sessionpb.CallbackBehaviour(v.Behaviour),
		}
	}

	pInputs := make(map[string]string, len(s.PendingInputs))
	for k, v := range s.PendingInputs {
		if v == "" {
			continue
		}
		pInputs[k] = string(v)
	}

	return proto.Marshal(&sessionpb.Session{
		ExpireTime:       timestamppb.New(s.ExpireTime),
		Ttl:              s.TTL,
		UserId:           s.UserID,
		CurrentStage:     string(s.CurrentStage),
		Context:          pbCtx,
		PendingCallbacks: pbCbs,
		PendingInputs:    pInputs,
		StageMessages:    s.StageMessages,
	})
}

func (s *Session) UnmarshalProto(data []byte) error {
	var ps sessionpb.Session
	if err := proto.Unmarshal(data, &ps); err != nil {
		return err
	}

	if ps.ExpireTime != nil {
		s.ExpireTime = ps.ExpireTime.AsTime()
	} else {
		s.ExpireTime = time.Time{}
	}

	s.TTL = ps.Ttl
	s.UserID = ps.UserId
	s.CurrentStage = model.ResourceRef(ps.CurrentStage)

	var ctx *model.UserContext
	if ps.Context != nil {
		ctx = &model.UserContext{
			UserID:   ps.Context.UserId,
			Lang:     ps.Context.Lang,
			Name:     ps.Context.Name,
			LastName: ps.Context.LastName,
			Username: ps.Context.Username,
		}
		if ps.Context.Misc != nil {
			ctx.Misc = ps.Context.Misc.AsMap()
		} else {
			ctx.Misc = make(map[string]interface{})
		}
	}
	s.UserContext = ctx

	if len(ps.PendingCallbacks) > 0 {
		s.PendingCallbacks = make(map[string]*model.PendingCallback, len(ps.PendingCallbacks))
		for k, v := range ps.PendingCallbacks {
			s.PendingCallbacks[k] = &model.PendingCallback{
				HandlerRef: model.ResourceRef(v.HandlerRef),
				Behaviour:  model.CallbackBehaviour(v.Behaviour),
			}
		}
	} else {
		s.PendingCallbacks = make(map[string]*model.PendingCallback)
	}

	if len(ps.PendingInputs) > 0 {
		s.PendingInputs = make(map[string]model.ResourceRef, len(ps.PendingInputs))
		for k, v := range ps.PendingInputs {
			s.PendingInputs[k] = model.ResourceRef(v)
		}
	} else {
		s.PendingInputs = make(map[string]model.ResourceRef)
	}

	s.StageMessages = ps.StageMessages
	return nil
}
