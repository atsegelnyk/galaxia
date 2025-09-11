package model

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type UserActionFunc func(ctx *UserContext, update *tgbotapi.Update) Updater

type Action struct {
	name string
	fn   UserActionFunc
}

func NewAction(name string, actionFunc UserActionFunc) *Action {
	return &Action{
		name: name,
		fn:   actionFunc,
	}
}

func (s *Action) SelfRef() ResourceRef {
	return ResourceRef(s.name)
}

func (s *Action) Func() UserActionFunc {
	return s.fn
}
