package auth

import "errors"

var UnauthorizedErr = errors.New("unauthorized user")

type Auther interface {
	AuthN(userID int64) error
}

type FakeAlwaysAuther struct {
}

func NewFakeAlwaysAuther() *FakeAlwaysAuther {
	return &FakeAlwaysAuther{}
}

func (f *FakeAlwaysAuther) AuthN(_ int64) error {
	return nil
}
