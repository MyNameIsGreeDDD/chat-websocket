package receivers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/go-playground/validator/v10"
)

type (
	wsServiceInterface interface {
		WriteServerBinary(msg []byte, conn net.Conn) error
	}

	MessageReadReceiver struct {
		Message   []byte
		Conn      net.Conn
		wsService wsServiceInterface
	}

	MessageReadEvent struct {
		Event  string          `json:"event" validate:"required"`
		UserId int             `json:"user_id" validate:"required"`
		Data   MessageReadData `json:"data" validate:"required"`
		Socket string          `json:"socket"`
	}
	MessageReadData struct {
		InfoForClient InfoForClient `json:"info_for_client" validate:"required"`
	}
	InfoForClient struct {
		ChatId     int `json:"chat_id" validate:"required"`
		MessageId  int `json:"message_id" validate:"required"`
		ReadUserId int `json:"read_user_id" validate:"required"`
	}
)

func NewMessageReadReceiver(wsService wsServiceInterface) *MessageReadReceiver {
	return &MessageReadReceiver{
		wsService: wsService,
	}
}

func (m *MessageReadReceiver) Receive(msg []byte, conn net.Conn) error {
	messageReadEvent := &MessageReadEvent{}
	err := json.Unmarshal(msg, messageReadEvent)
	if err != nil {
		return errors.New(fmt.Sprintf("cant unmarshal %s", err))
	}

	err = validator.New().Struct(messageReadEvent)
	if err != nil {
		return errors.New(fmt.Sprintf("something wrong while write message in worker: %s", err))
	}

	err = m.wsService.WriteServerBinary(msg, conn)
	if err != nil {
		return errors.New(fmt.Sprintf("something wrong while write message in worker: %s", err))
	}

	return nil
}
