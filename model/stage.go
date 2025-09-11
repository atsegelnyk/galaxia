package model

import (
	"errors"
)

var UnrecognizedInputError = errors.New("unresolved input")

type PendingInput struct {
	Body      string      `json:"body,omitempty"`
	ActionRef ResourceRef `json:"action_ref" json:"action_ref,omitempty"`
}

type Stage struct {
	name               string
	customInputAllowed bool

	initializer   StageInitializer
	defaultAction ResourceRef
}

type StageOption func(*Stage)

func NewStage(name string, opts ...StageOption) *Stage {
	s := &Stage{
		name: name,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithInitializer(initializer StageInitializer) StageOption {
	return func(stage *Stage) {
		stage.initializer = initializer
	}
}

func (s *Stage) WithInitializer(initializer StageInitializer) *Stage {
	s.initializer = initializer
	return s
}

func WithCustomInputAllowed(allowed bool) StageOption {
	return func(stage *Stage) {
		stage.customInputAllowed = allowed
	}
}

func (s *Stage) WithCustomInputAllowed(allowed bool) *Stage {
	s.customInputAllowed = allowed
	return s
}

func WithDefaultAction(defaultAction ResourceRef) StageOption {
	return func(stage *Stage) {
		stage.defaultAction = defaultAction
	}
}

func (s *Stage) WithDefaultAction(defaultAction ResourceRef) *Stage {
	s.defaultAction = defaultAction
	return s
}

func (s *Stage) SelfRef() ResourceRef {
	return ResourceRef(s.name)
}

func (s *Stage) Initializer() StageInitializer {
	if s.initializer != nil {
		return s.initializer
	}
	return nil
}

func (s *Stage) DefaultActionRef() ResourceRef {
	return s.defaultAction
}

func (s *Stage) CustomInputAllowed() bool {
	return s.customInputAllowed
}

func (s *Stage) Initialize(userID int64) (*Message, error) {
	return s.initializer.Get(userID)
}

// StageInitializer represents initializer interface
type StageInitializer interface {
	Get(userId int64) (*Message, error)
}

type StaticStageInitializer struct {
	Message *Message
}

func NewStaticStageInitializer(msg *Message) *StaticStageInitializer {
	return &StaticStageInitializer{
		Message: msg,
	}
}

func (s *StaticStageInitializer) Get(_ int64) (*Message, error) {
	return s.Message, nil
}
