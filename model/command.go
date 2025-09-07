package model

type Command struct {
	name    string
	handler UserAction
}

func NewCommand(name string, handler UserAction) *Command {
	return &Command{
		name:    name,
		handler: handler,
	}
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Handler() UserAction {
	return c.handler
}
