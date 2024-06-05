package read

import (
	"context"
	"github.com/redis/go-redis/v9"
	"net"
	"strconv"
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

func (r *RedisReadRepository) GetConnectionByUserId(userId int) (*net.Conn, error) {
	var userConn UserConn

	_, err := (*r.client).Get(context.Background(), strconv.Itoa(userId)).Result()
	if err != nil {
		return nil, err
	}

	return &userConn.connection, nil
}

func (r *RedisReadRepository) HGet(key, field string) adapters.StringCmdInterface {
	return (*r.client).HGet(context.Background(), key, field)
}
