package publish

import (
	"errors"
	"fmt"

	"websocket-confee/internal/adapters"

	e "websocket-confee/internal/services/event"
)

type Director struct {
	publisher e.Worker
}

type redisServiceInterface interface {
	Publish(message interface{}) error
	HGet(key string, field string) adapters.StringCmdInterface
}

func NewPublishDirector(service redisServiceInterface, event string, msg []byte) (*Director, error) {
	publisher, ok := e.AllowedPubs[event]
	if !ok {
		return nil, errors.New(fmt.Sprintf("event: %s doesnt support", event))
	}

	return &Director{
		publisher: publisher(msg, service),
	}, nil
}

func (d *Director) Run() error {
	return d.publisher.Run()
}
