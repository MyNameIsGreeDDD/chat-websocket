package write

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisWriteRepository struct {
	client *redis.Client
}

func NewRedisWriteRepository(client *redis.Client) *RedisWriteRepository {
	return &RedisWriteRepository{
		client: client,
	}
}

func (w *RedisWriteRepository) Publish(message []byte, channel string) error {
	return (*w.client).Publish(context.Background(), channel, message).Err()
}

func (w *RedisWriteRepository) Store(userId int, conn *net.Conn) error {
	c, err := json.Marshal(conn)
	if err != nil {
		return err
	}

	_, err = (*w.client).Set(context.Background(), strconv.Itoa(userId), c, 7*24*time.Hour).Result()
	return err
}
