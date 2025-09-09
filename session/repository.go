package session

import (
	"errors"
)

var NotFoundError = errors.New("session not found")

type Repository interface {
	Get(userID int64) (*Session, error)
	Save(session *Session) error
	Expire(userID int64)
}
