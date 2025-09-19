package model

type PendingCallback struct {
	UserData   string            `json:"userData,omitempty"`
	HandlerRef ResourceRef       `json:"handler_ref" json:"handlerRef,omitempty"`
	Behaviour  CallbackBehaviour `json:"behaviour" json:"behaviour,omitempty"`
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
