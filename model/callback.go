package model

type PendingCallback struct {
	HandlerRef ResourceRef       `json:"handler_ref"`
	Behaviour  CallbackBehaviour `json:"behaviour"`
}

type CallbackHandler struct {
	name      string
	actionRef ResourceRef
}

func NewCallbackHandler(name string, actionRef ResourceRef) *CallbackHandler {
	return &CallbackHandler{
		name:      name,
		actionRef: actionRef,
	}
}

func (c *CallbackHandler) SelfRef() ResourceRef {
	return ResourceRef(c.name)
}

func (c *CallbackHandler) ActionRef() ResourceRef {
	return c.actionRef
}
