package model

type ResourceRef string

type Referencer interface {
	SelfRef() ResourceRef
}
