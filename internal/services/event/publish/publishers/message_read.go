package publishers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type redisService interface {
	Publish(message interface{}) error
}

type (
	MessageRead struct {
		Message      []byte
		RedisService redisService
	}
	MessageReadEvent struct {
		Event  string          `json:"event" validate:"required"`
		UserId int             `json:"userId" validate:"required"`
		Data   MessageReadData `json:"data" validate:"required"`
	}
	MessageReadData struct {
		ChatId    int    `json:"chat_id" validate:"required"`
		MessageId int    `json:"message_id" validate:"required"`
		SessionId string `json:"session_id" validate:"required"`
	}
)

func NewMessageReadPublisher(redisService redisService) *MessageRead {
	return &MessageRead{
		RedisService: redisService,
	}
}

func (m *MessageRead) Publish(msg []byte) error {
	messageReadEvent := &MessageReadEvent{}

	err := json.Unmarshal(msg, &messageReadEvent)
	if err != nil {
		return errors.New(fmt.Sprintf("cant unmarshal %s", err))
	}

	err = validator.New().Struct(messageReadEvent)
	if err != nil {
		return errors.New(fmt.Sprintf("something wrong while write message in worker: %s", err))
	}

	err = m.RedisService.Publish(messageReadEvent)
	if err != nil {
		return errors.New(fmt.Sprintf("cant publish event %s", err))
	}

	return nil
}
