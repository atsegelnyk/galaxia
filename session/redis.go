package session

// TODO implement SessionRepository
type RedisSessionRepository struct {
}

func NewRedisSessionRepository() *RedisSessionRepository {
	return &RedisSessionRepository{}
}
