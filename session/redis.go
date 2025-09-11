package session

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const sessionKey = "%d:session"

type RedisSessionRepository struct {
	ctx           context.Context
	keyPrefix     string
	client        *redis.Client
	clusterClient *redis.ClusterClient
}

type RedisSessionRepositoryOption func(*RedisSessionRepository)

func NewRedisSessionRepository(ctx context.Context, opts ...RedisSessionRepositoryOption) *RedisSessionRepository {
	r := &RedisSessionRepository{
		ctx: ctx,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithClient(client *redis.Client) RedisSessionRepositoryOption {
	return func(r *RedisSessionRepository) {
		r.client = client
	}
}

func WithClusterClient(client *redis.ClusterClient) RedisSessionRepositoryOption {
	return func(r *RedisSessionRepository) {
		r.clusterClient = client
	}
}

func WithKeyPrefix(prefix string) RedisSessionRepositoryOption {
	return func(r *RedisSessionRepository) {
		r.keyPrefix = prefix
	}
}

func (r *RedisSessionRepository) Get(userID int64) (*Session, error) {
	sessionResponse := r.get(r.buildSessionKey(userID))
	_, err := sessionResponse.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, NotFoundError
		}
		return nil, err
	}

	sessionBytes, err := sessionResponse.Bytes()
	if err != nil {
		return nil, err
	}

	ses := &Session{}
	err = ses.UnmarshalProto(sessionBytes)
	if err != nil {
		return nil, err
	}

	return ses, nil
}

func (r *RedisSessionRepository) Save(session *Session) error {
	sessionBytes, err := session.MarshalProto()
	if err != nil {
		return err
	}

	setResp := r.set(
		r.buildSessionKey(session.UserID),
		sessionBytes,
		time.Duration(session.TTL)*time.Second,
	)

	_, err = setResp.Result()
	return err
}

func (r *RedisSessionRepository) Expire(userID int64) {
	_ = r.delete(r.buildSessionKey(userID))
}

func (r *RedisSessionRepository) get(key string) *redis.StringCmd {
	if r.client != nil {
		return r.client.Get(r.ctx, key)
	}
	return r.clusterClient.Get(r.ctx, key)
}

func (r *RedisSessionRepository) delete(key string) *redis.IntCmd {
	if r.client != nil {
		return r.client.Del(r.ctx, key)
	}
	return r.clusterClient.Del(r.ctx, key)
}

func (r *RedisSessionRepository) set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if r.client != nil {
		return r.client.Set(r.ctx, key, value, expiration)
	}
	return r.clusterClient.Set(r.ctx, key, value, expiration)
}

func (r *RedisSessionRepository) buildSessionKey(userID int64) string {
	keyBase := fmt.Sprintf(sessionKey, userID)
	if r.keyPrefix != "" {
		keyBase = r.keyPrefix + keyBase
	}
	return keyBase
}
