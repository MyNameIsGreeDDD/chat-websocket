package event

import (
	"net"

	"websocket-confee/internal/adapters"
	"websocket-confee/internal/models"
	"websocket-confee/internal/services/event/publish/publishers"
	"websocket-confee/internal/services/event/receive/receivers"
)

type (
	Worker interface {
		Run() error
	}

	redisInterface interface {
		Publish(message interface{}) error
		HGet(key, field string) adapters.StringCmdInterface
	}
	wsServiceInterface interface {
		WriteServerBinary(msg []byte, conn net.Conn) error
	}

	CreatePubFn func(
		message []byte,
		service redisInterface,
	) Worker

	CreateSubFn func(
		message []byte,
		conn net.Conn,
		wsService wsServiceInterface,
	) Worker
)

var (
	AllowedPubs = make(map[string]CreatePubFn, 1)
	AllowedSubs = make(map[string]CreateSubFn, 1)
)

func init() {
	AllowedSubs[models.MessageRead] = func(message []byte, conn net.Conn, wsService wsServiceInterface) Worker {
		return receivers.NewMessageReadReceiver(message, conn, wsService)
	}
	AllowedPubs[models.MessageRead] = func(message []byte, service redisInterface) Worker {
		return publishers.NewMessageReadPublisher(message, service)
	}
}
