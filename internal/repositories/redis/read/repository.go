package read

import (
	"context"
	"net"

	"github.com/redis/go-redis/v9"

	"websocket-confee/internal/adapters"
)

type UserConn struct {
	connection net.Conn
}

type RedisReadRepository struct {
	client *redis.Client
}

func NewRedisReadRepository(client *redis.Client) *RedisReadRepository {
	return &RedisReadRepository{
		client: client,
	}
}

func (r *RedisReadRepository) Subscribe(channel ...string) adapters.PubSubInterface {
	return (*r.client).Subscribe(context.Background(), channel...)
}
func (r *RedisReadRepository) HGet(key, field string) adapters.StringCmdInterface {
	return (*r.client).HGet(context.Background(), key, field)
}
func (w *RedisReadRepository) GetUserIdBySessionId(sessionId string) (int, error) {
	return (*w.client).Get(context.Background(), sessionId).Int()
}
