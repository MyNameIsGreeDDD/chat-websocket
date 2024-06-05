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
	"websocket-confee/internal/services/event/publish"
)

type wsServiceInterface interface {
	WriteClientBinary(msg []byte, conn net.Conn) error
	WriteClientClose(msg []byte, conn net.Conn) error
	ReadClientMessage(reader adapters.ReaderInterface) ([]byte, error)
	NewReader(conn net.Conn) adapters.ReaderInterface
}

type loggerInterface interface {
	Error(err string)
}

type CopyHandler struct {
	redis        redisServiceInterface
	connections  map[int]net.Conn
	conn         net.Conn
	connWg       *sync.WaitGroup
	connPool     chan struct{}
	logger       loggerInterface
	wsService    wsServiceInterface
	rwConnsMutex *sync.RWMutex
	ctx          context.Context
}

func NewHandler(
	redisService redisServiceInterface,
	wsService wsServiceInterface,
	logger loggerInterface,
	connections map[int]net.Conn,
	connPool chan struct{},
	connWg *sync.WaitGroup,
	conn net.Conn,
	connMutex *sync.RWMutex,
	ctx context.Context,
) *CopyHandler {
	return &CopyHandler{
		redis:        redisService,
		connections:  connections,
		conn:         conn,
		connWg:       connWg,
		connPool:     connPool,
		logger:       logger,
		wsService:    wsService,
		rwConnsMutex: connMutex,
		ctx:          ctx,
	}
}

func (h *CopyHandler) Handle() {
	h.connWg.Add(1)
	h.connPool <- struct{}{}

	go func() {
		defer h.connWg.Done()
		defer func() { <-h.connPool }()
		defer func() {
			if r := recover(); r != nil {
				h.handleError(errors.New(fmt.Sprintf(" Handled panic: %v/n", r)))
			}
		}()

		msgWg := &sync.WaitGroup{}
		msgHandlePool := make(chan struct{}, 2)

		rd := h.wsService.NewReader(h.conn)

		for {
			msg, err := h.wsService.ReadClientMessage(rd)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					// Эта ошибка может появится только при установленном ReadDeadLine. Она лишь означает что
					// в установленное время не произашла запись. Нам эта ошибка нужна,
					// чтобы тригерить проверку ниже, срабатывающией при падении приложения. Если этого не будет,
					// то graceful shutdown(gc) будет крутиться вечно для каждого обработчика сообщения и gc не закончится.

					select {
					case <-h.ctx.Done():
						return
					default:
						continue
					}
				}

				h.handleError(err)
				continue
			}

			msgHandlePool <- struct{}{}
			msgWg.Add(1)

			go func(msg []byte) {
				defer msgWg.Done()
				defer func() { <-msgHandlePool }()
				defer func() {
					if r := recover(); r != nil {
						h.handleError(errors.New(fmt.Sprintf(" Handled panic: %v/n", r)))
					}
				}()
				baseEvent := models.BaseEvent{}
				if err := json.Unmarshal(msg, &baseEvent); err != nil {
					h.handleError(err)
				}

				switch baseEvent.Event {
				case models.Auth:
					err = NewAuthHandler(h.redis, h.wsService, h.connections, h.rwConnsMutex, h.conn, msg).Handle()
					if err != nil {
						h.logger.Error(errors.New("403 forbidden").Error())
						return
					}
				default:
					publisher, err := publish.NewPublishDirector(h.redis, baseEvent.Event, msg)
					if err != nil {
						h.logger.Error(errors.New(fmt.Sprintf("something wrong while create pub director %s", err.Error())).Error())
						return
					}

					if err = publisher.Run(); err != nil {
						h.logger.Error(errors.New(fmt.Sprintf("something wrong while publishing message %s", err.Error())).Error())
						return
					}
				}
			}(msg)
		}
		msgWg.Wait()
	}()
}

func (h *CopyHandler) handleError(err error) {
	h.wsService.WriteClientClose([]byte("connection closed"), h.conn)
	h.conn.Close()
	h.logger.Error(fmt.Sprintf("failed handle message with error: %s", err.Error()))
}
