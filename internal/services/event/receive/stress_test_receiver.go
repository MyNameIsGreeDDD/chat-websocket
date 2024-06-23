package receive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"websocket-confee/internal/models"
	"websocket-confee/internal/services/event"
)

type StressTestEventSub struct {
	RedisService       redisServiceInterface
	log                loggerInterface
	wsService          wsServiceInterface
	connections        map[int]net.Conn
	subsPool           chan struct{}
	subsWg             *sync.WaitGroup
	ctx                context.Context
	allowedReceivers   map[string]event.Receiver
	mockMessageChannel chan []byte
	connMutex          *sync.RWMutex
	sync.RWMutex
}

func RegisterStressTestEventListener(
	RedisService redisServiceInterface,
	wsService wsServiceInterface,
	log loggerInterface,
	connections map[int]net.Conn,
	subsPool chan struct{},
	subsWg *sync.WaitGroup,
	ctx context.Context,
	allowedReceivers map[string]event.Receiver,
	mockMessageChannel chan []byte,
	connMutex *sync.RWMutex,
) *StressTestEventSub {
	return &StressTestEventSub{
		RedisService:       RedisService,
		connections:        connections,
		subsWg:             subsWg,
		log:                log,
		subsPool:           subsPool,
		wsService:          wsService,
		ctx:                ctx,
		allowedReceivers:   allowedReceivers,
		mockMessageChannel: mockMessageChannel,
		connMutex:          connMutex,
	}
}

func (e *StressTestEventSub) Run() error {
	//defaultEvents := e.RedisService.Subscribe(r.SubEventChannel)

	go func(subsWg *sync.WaitGroup, subsPool chan struct{}, log loggerInterface) {
		for msg := range e.mockMessageChannel {
			select {
			case <-e.ctx.Done():
				subsPool <- struct{}{}
				subsWg.Add(1)

				go e.handleMessage(msg)
				return
			default:
				subsPool <- struct{}{}
				subsWg.Add(1)

				go e.handleMessage(msg)
			}
		}
	}(e.subsWg, e.subsPool, e.log)

	return nil
}

func (e *StressTestEventSub) handleMessage(msg []byte) {
	defer func() {
		<-e.subsPool
		e.subsWg.Done()
		fmt.Println("sub pool" + strconv.Itoa(len(e.subsPool)))
	}()
	defer func(log loggerInterface, msg []byte) {
		if rec := recover(); rec != nil {
			log.Error(errors.New(fmt.Sprintf(" Handled panic: %v/n with message: %s", rec, string(msg))).Error())
		}
	}(e.log, msg)

	baseEvent := &models.BaseSubEvent{}
	if err := json.Unmarshal(msg, &baseEvent); err != nil {
		e.log.Error(errors.New(fmt.Sprintf("cant unmarshal message: %v", msg)).Error())
	}

	receiver, ok := e.allowedReceivers[baseEvent.Event]
	if !ok {
		e.log.Error(errors.New(fmt.Sprintf("event: %s doesnt support", baseEvent.Event)).Error())
	}

	e.connMutex.RLock()
	conn, ok := e.connections[baseEvent.UserId]
	e.connMutex.RUnlock()
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

	err := receiver.Receive(msg, conn)
	if err != nil {
		e.log.Error(errors.New("handled error while sending message").Error())
	}
}
