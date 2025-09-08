package model

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type CallbackHandlerFunc func(update *tgbotapi.Update) Responser

type CallbackContext struct {
	Retain  bool
	Handler *CallbackHandler
	Misc    interface{}
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
