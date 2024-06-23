package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"websocket-confee/internal/adapters"
	"websocket-confee/internal/models"
	"websocket-confee/internal/services/event"
)

type (
	wsServiceInterface interface {
		WriteServerBinary(msg []byte, conn net.Conn) error
		WriteServerClose(msg []byte, conn net.Conn) error
		ReadClientMessage(reader adapters.ReaderInterface) ([]byte, error)
		NewReader(conn net.Conn) adapters.ReaderInterface
	}
	redisServiceInterface interface {
		HGet(key, field string) adapters.StringCmdInterface
		GetUserIdBySessionId(sessionId string) (int, error)
		Publish(message interface{}) error
		StoreSessionId(sessionId string, userId int) error
	}

	loggerInterface interface {
		Error(err string)
	}

	AuthEvent struct {
		Event string   `json:"event" validate:"required"`
		Data  AuthData `json:"data" validate:"required"`
	}
	AuthData struct {
		Token string `json:"token" validate:"required"`
	}
	ConnectedEvent struct {
		Event string        `json:"event" validate:"required"`
		Data  ConnectedData `json:"data" validate:"required"`
	}
	ConnectedData struct {
		SessionID string `json:"session_id" validate:"required"`
	}

	Handler struct {
		redis             redisServiceInterface
		connections       map[int]net.Conn
		connWg            *sync.WaitGroup
		connPool          chan struct{}
		logger            loggerInterface
		wsService         wsServiceInterface
		rwConnsMutex      *sync.RWMutex
		ctx               context.Context
		allowedPublishers map[string]event.Publisher
	}
)

func NewHandler(
	redisService redisServiceInterface,
	wsService wsServiceInterface,
	logger loggerInterface,
	connections map[int]net.Conn,
	connPool chan struct{},
	connWg *sync.WaitGroup,
	connMutex *sync.RWMutex,
	ctx context.Context,
	allowedPublishers map[string]event.Publisher,
) *Handler {
	return &Handler{
		redis:             redisService,
		connections:       connections,
		connWg:            connWg,
		connPool:          connPool,
		logger:            logger,
		wsService:         wsService,
		rwConnsMutex:      connMutex,
		ctx:               ctx,
		allowedPublishers: allowedPublishers,
	}
}

func (h *Handler) Handle(conn net.Conn) {
	h.connWg.Add(1)
	h.connPool <- struct{}{}

	go func() {
		defer h.connWg.Done()
		defer func() { <-h.connPool }()
		defer func() {
			if r := recover(); r != nil {
				h.handleError(errors.New(fmt.Sprintf(" Handled panic: %v/n", r)), conn)
			}
		}()

		msgWg := &sync.WaitGroup{}
		msgHandlePool := make(chan struct{}, 2)

		rd := h.wsService.NewReader(conn)

		for {
			msg, err := h.wsService.ReadClientMessage(rd)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					// Эта ошибка может появится только при установленном ReadDeadLine. Она лишь означает что
					// в установленное время не произашла запись. Нам эта ошибка нужна,
					// чтобы тригерить проверку ниже, срабатывающией при падении приложения. Если этого не будет,
					// то graceful shutdown(gs) будет крутиться вечно для каждого обработчика сообщения и gs не закончится.

					select {
					case <-h.ctx.Done():
						return
					default:
						continue
					}
				}

				h.handleError(err, conn)
				continue
			}

			msgHandlePool <- struct{}{}
			msgWg.Add(1)

			go func(msg []byte) {
				defer msgWg.Done()
				defer func() { <-msgHandlePool }()
				defer func() {
					if r := recover(); r != nil {
						h.handleError(errors.New(fmt.Sprintf(" Handled panic: %v/n", r)), conn)
					}
				}()
				baseEvent := models.BaseEvent{}
				if err := json.Unmarshal(msg, &baseEvent); err != nil {
					h.handleError(err, conn)
				}

				switch baseEvent.Event {
				case models.Auth:
					err = h.handleAuth(msg, conn)
					if err != nil {
						h.logger.Error(errors.New("403 forbidden").Error())
						return
					}
				default:
					_, err := h.redis.GetUserIdBySessionId(baseEvent.SessionId)
					if err != nil {
						h.handleError(errors.New("403 forbidden"), conn)
						return
					}

					publisher, ok := h.allowedPublishers[baseEvent.Event]
					if !ok {
						h.logger.Error(errors.New(fmt.Sprintf("event: %s doesnt support", baseEvent.Event)).Error())
						return
					}

					if err = publisher.Publish(msg); err != nil {
						h.logger.Error(errors.New(fmt.Sprintf("something wrong while publishing message %s", err.Error())).Error())
						return
					}
				}
			}(msg)
		}
		msgWg.Wait()
	}()
}

func (h *Handler) handleAuth(msg []byte, conn net.Conn) error {
	authEvent := &AuthEvent{}

	err := json.Unmarshal(msg, authEvent)
	if err != nil {
		return errors.New(fmt.Sprintf("cant unmarhsal %s", err))
	}

	userId, err := h.redis.HGet(authEvent.Data.Token, "user_id").Int()
	sessionId := h.redis.HGet(authEvent.Data.Token, "session_id").String()
	if err != nil || sessionId == "" {
		return errors.New(fmt.Sprintf("403 forbidden %s", err))
	}

	h.rwConnsMutex.Lock()
	h.connections[userId] = conn
	h.rwConnsMutex.Unlock()

	err = h.redis.StoreSessionId(sessionId, userId)
	if err != nil {
		return errors.New(fmt.Sprintf("cant store session id in handler %s", err))
	}

	err = h.redis.Publish(&ConnectedEvent{
		Event: models.Connected,
		Data: ConnectedData{
			SessionID: sessionId,
		},
	})
	if err != nil {
		return errors.New(fmt.Sprintf("something wrong while write connected event in redis %s", err))
	}

	successResponse, _ := json.Marshal(models.SuccessResponse{
		Message: "auth success",
		Code:    200,
	})

	err = h.wsService.WriteServerBinary(successResponse, conn)
	if err != nil {
		return errors.New(fmt.Sprintf("something wrong while write message in worker: %s", err))
	}

	return nil
}

func (h *Handler) handleError(err error, conn net.Conn) {
	if conn != nil {
		h.wsService.WriteServerClose([]byte("connection closed"), conn)
		conn.Close()
	}

	h.logger.Error(fmt.Sprintf("failed handle message with error: %s", err.Error()))
}
