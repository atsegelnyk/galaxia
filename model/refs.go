package model

type ResourceRef string

func (r ResourceRef) Empty() bool {
	return r == ""
}

type Referencer interface {
	SelfRef() ResourceRef
}
