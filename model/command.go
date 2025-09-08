package model

type Command struct {
	name    string
	handler UserActionFunc
}

func NewCommand(name string, handler UserActionFunc) *Command {
	return &Command{
		name:    name,
		handler: handler,
	}
}

func (c *Command) SelfRef() ResourceRef {
	return ResourceRef(c.name)
}

func (c *Command) Handler() UserActionFunc {
	return c.handler
}
