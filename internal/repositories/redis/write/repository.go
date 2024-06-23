package write

import (
	"context"
	"errors"
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

func (w *RedisWriteRepository) StoreSessionId(sessionId string, userId int) error {
	err := (*w.client).Set(context.Background(), strconv.Itoa(userId), sessionId, 24*time.Hour).Err()
	if err != nil {
		return errors.New("failed store session_id for user_id: " + strconv.Itoa(userId))
	}

	return nil
}
