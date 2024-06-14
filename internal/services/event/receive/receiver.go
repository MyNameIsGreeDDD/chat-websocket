package receive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"websocket-confee/internal/adapters"
	"websocket-confee/internal/models"
	"websocket-confee/internal/services/event"
	r "websocket-confee/internal/services/redis"
)

type redisServiceInterface interface {
	Subscribe(channel ...string) adapters.PubSubInterface
}
type wsServiceInterface interface {
	WriteServerBinary(msg []byte, conn net.Conn) error
}
type loggerInterface interface {
	Error(msg string)
	Info(msg string)
}

type EventSub struct {
	RedisService redisServiceInterface
	log          loggerInterface
	wsService    wsServiceInterface
	connections  map[int]net.Conn
	subsPool     chan struct{}
	subsWg       *sync.WaitGroup
	ctx          context.Context
}

func RegisterEventListener(
	RedisService redisServiceInterface,
	wsService wsServiceInterface,
	log loggerInterface,
	connections map[int]net.Conn,
	subsPool chan struct{},
	subsWg *sync.WaitGroup,
	ctx context.Context,
) event.Worker {
	return &EventSub{
		RedisService: RedisService,
		connections:  connections,
		subsWg:       subsWg,
		log:          log,
		subsPool:     subsPool,
		wsService:    wsService,
		ctx:          ctx,
	}
}

func (e *EventSub) Run() error {
	defaultEvents := e.RedisService.Subscribe(r.SubEventChannel)

	go func(subsWg *sync.WaitGroup, subsPool chan struct{}, log loggerInterface) {
		for msg := range defaultEvents.Channel() {
			select {
			case <-e.ctx.Done():
				return
			default:
				subsPool <- struct{}{}
				subsWg.Add(1)

				go e.handleMessage([]byte(msg.Payload))
			}
		}
	}(e.subsWg, e.subsPool, e.log)

	return nil
}

func (e *EventSub) handleMessage(msg []byte) {
	defer e.subsWg.Done()
	defer func() { <-e.subsPool }()
	defer func(log loggerInterface, msg []byte) {
		if rec := recover(); rec != nil {
			log.Error(errors.New(fmt.Sprintf(" Handled panic: %v/n with message: %s", rec, string(msg))).Error())
		}
	}(e.log, msg)

	baseEvent := &models.BaseSubEvent{}
	if err := json.Unmarshal(msg, &baseEvent); err != nil {
		e.log.Error(errors.New(fmt.Sprintf("cant unmarshal message: %v", msg)).Error())
	}

	sub, ok := event.AllowedSubs[baseEvent.Event]
	if !ok {
		e.log.Error(errors.New(fmt.Sprintf("event: %s doesnt support", baseEvent.Event)).Error())
	}

	conn, ok := e.connections[baseEvent.UserId]
	if !ok {
		e.log.Error(errors.New(fmt.Sprintf("403 forbidden for user: %d", baseEvent.UserId)).Error())
		return
	}

	if conn == nil {
		// В целом не нужно помещать сообщения обратно в случае если подключение было оборвано со стороны клиента
		// т.к клиент по какой-то причине не перестал удерживать подключение,
		// а значит - все данные он может получить при обновлении страницы.
		return
	}

	err := sub(msg, conn, e.wsService).Run()
	if err != nil {
		e.log.Error(errors.New("handled error while sending message").Error())
	}
}
