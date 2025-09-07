package auth

type Auther interface {
	Authorize(userID int64) error
}
