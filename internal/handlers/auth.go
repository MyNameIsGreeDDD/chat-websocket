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
		rwMutex     *sync.RWMutex
		connection  net.Conn
		msg         []byte
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
	conn net.Conn,
	msg []byte,
) *Auth {
	return &Auth{
		connections: connections,
		redis:       redis,
		msg:         msg,
		connection:  conn,
		rwMutex:     rwMutex,
		wsService:   wsService,
	}
}

func (a *Auth) Handle() error {
	event := &AuthEvent{}

	err := json.Unmarshal(a.msg, event)
	if err != nil {
		return errors.New(fmt.Sprintf("cant unmarhsal %s", err))
	}

	userId, err := a.redis.HGet(event.Data.Token, "user_id").Int()
	if err != nil {
		return errors.New(fmt.Sprintf("403 forbidden %s", err))
	}

	a.rwMutex.Lock()
	a.connections[userId] = a.connection
	a.rwMutex.Unlock()

	successResponse, _ := json.Marshal(models.SuccessResponse{
		Message: "auth success",
		Code:    200,
	})

	err = a.wsService.WriteServerBinary(successResponse, a.connection)
	if err != nil {
		return errors.New(fmt.Sprintf("something wrong while write message in worker: %s", err))
	}

	return nil
}
