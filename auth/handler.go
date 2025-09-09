package auth

import "errors"

var UnauthorizedErr = errors.New("unauthorized user")

type Auther interface {
	Authorize(userID int64) error
}
