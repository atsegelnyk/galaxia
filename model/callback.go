package model

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type CallbackHandlerFunc func(ctx *UserContext, update *tgbotapi.Update) Responser

type PendingCallback struct {
	HandlerRef ResourceRef
	Behaviour  CallbackBehaviour
}

type CallbackHandler struct {
	name    string
	handler CallbackHandlerFunc
}

func NewCallbackHandler(name string, handler CallbackHandlerFunc) *CallbackHandler {
	return &CallbackHandler{
		name:    name,
		handler: handler,
	}
}

func (c *CallbackHandler) SelfRef() ResourceRef {
	return ResourceRef(c.name)
}

func (c *CallbackHandler) Func() CallbackHandlerFunc {
	return c.handler
}
