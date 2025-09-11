package model

type Command struct {
	name      string
	actionRef ResourceRef
}

func NewCommand(name string, actionRef ResourceRef) *Command {
	return &Command{
		name:      name,
		actionRef: actionRef,
	}
}

func (c *Command) SelfRef() ResourceRef {
	return ResourceRef(c.name)
}

func (c *Command) ActionRef() ResourceRef {
	return c.actionRef
}
