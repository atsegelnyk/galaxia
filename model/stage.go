package model

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var UnsupportedInputError = errors.New("unsupported input")

type Stage struct {
	name         string
	inputAllowed bool

	initializer  StageInitializer
	inputHandler InputHandler
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

func WithInputHandler(handler InputHandler) StageOption {
	return func(stage *Stage) {
		stage.inputAllowed = true
		stage.inputHandler = handler
	}
}

func (s *Stage) Name() string {
	return s.name
}

func (s *Stage) Initializer() StageInitializer {
	if s.initializer != nil {
		return s.initializer
	}
	return nil
}

func (s *Stage) ProcessUserEvent(update *tgbotapi.Update) (Responser, error) {
	initialMessage, err := s.initializer.Get(update.Message.Chat.ID)
	if err != nil {
		return nil, err
	}

	for _, r := range initialMessage.ReplyKeyboard {
		for _, b := range r {
			if b.Text == update.Message.Text {
				responser := b.Action(update)
				return responser, nil
			}
		}
	}
	if !s.inputAllowed {
		return nil, UnsupportedInputError
	}
	responser := s.inputHandler(update)
	return responser, nil
}

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
