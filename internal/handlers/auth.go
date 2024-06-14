package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"websocket-confee/internal/adapters"
	"websocket-confee/internal/models"
)

type (
	redisServiceInterface interface {
		HGet(key, field string) adapters.StringCmdInterface
		Publish(message interface{}) error
	}
	Auth struct {
		redis       redisServiceInterface
		wsService   wsServiceInterface
		connections map[int]net.Conn
		connMutex   *sync.RWMutex
	}
	AuthEvent struct {
		Event string   `json:"event" validate:"required"`
		Data  AuthData `json:"data" validate:"required"`
	}
	AuthData struct {
		Token string `json:"token" validate:"required"`
	}
)

func NewAuthHandler(
	redis redisServiceInterface,
	wsService wsServiceInterface,
	connections map[int]net.Conn,
	rwMutex *sync.RWMutex,

) *Auth {
	return &Auth{
		connections: connections,
		redis:       redis,
		connMutex:   rwMutex,
		wsService:   wsService,
	}
}

func (a *Auth) Handle(conn net.Conn, msg []byte) error {
	event := &AuthEvent{}

	err := json.Unmarshal(msg, event)
	if err != nil {
		return errors.New(fmt.Sprintf("cant unmarhsal %s", err))
	}

	userId, err := a.redis.HGet(event.Data.Token, "user_id").Int()
	if err != nil {
		return errors.New(fmt.Sprintf("403 forbidden %s", err))
	}

	a.connMutex.Lock()
	a.connections[userId] = conn
	a.connMutex.Unlock()

	successResponse, _ := json.Marshal(models.SuccessResponse{
		Message: "auth success",
		Code:    200,
	})

	err = a.wsService.WriteServerBinary(successResponse, conn)
	if err != nil {
		return errors.New(fmt.Sprintf("something wrong while write message in worker: %s", err))
	}

	return nil
}
