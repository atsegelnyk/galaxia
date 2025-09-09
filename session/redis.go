package session

// TODO implement SessionRepository
type RedisSessionRepository struct {
}

func NewRedisSessionRepository() *RedisSessionRepository {
	return &RedisSessionRepository{}
}

func (r *RedisSessionRepository) Get(userID int64) (*Session, error) {
	return nil, NotFoundError
}

func (r *RedisSessionRepository) Save(session *Session) error {
	return nil
}

func (r *RedisSessionRepository) Expire(userID int64) {

}
