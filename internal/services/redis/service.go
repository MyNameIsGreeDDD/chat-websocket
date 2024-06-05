package redis

import (
	"encoding/json"
	"net"
	"os"
	"websocket-confee/internal/adapters"
)

var PubEventChannel = os.Getenv("PUB_EVENTS_REDIS_CHANNEL")
var SubEventChannel = os.Getenv("SUB_EVENTS_REDIS_CHANNEL")

type writeRepositoryInterface interface {
	Publish(message []byte, channel string) error
	Store(userId int, conn *net.Conn) error
}

type readRepositoryInterface interface {
	Subscribe(channel ...string) adapters.PubSubInterface
	GetConnectionByUserId(userId int) (*net.Conn, error)
	HGet(key, field string) adapters.StringCmdInterface
}

type loggerInterface interface {
	Error(msg string)
	Info(msg string)
}

type Service struct {
	writeRepository writeRepositoryInterface
	readRepository  readRepositoryInterface
	logger          loggerInterface
}

func NewRedisService(
	logger loggerInterface,
	writeRepository writeRepositoryInterface,
	repositoryInterface readRepositoryInterface,
) *Service {
	return &Service{
		writeRepository: writeRepository,
		readRepository:  repositoryInterface,
		logger:          logger,
	}
}

func (s *Service) Publish(message interface{}) error {
	bytes, err := json.Marshal(message)

	if err != nil {
		s.logger.Error("something wrong while marshaling: " + err.Error())
		return err
	}

	return s.writeRepository.Publish(bytes, PubEventChannel)
}

func (s *Service) HGet(key, field string) adapters.StringCmdInterface {
	return s.readRepository.HGet(key, field)
}

func (s *Service) Subscribe(channel ...string) adapters.PubSubInterface {
	return s.readRepository.Subscribe(channel...)
}

func (s *Service) GetConnectionByUserId(userId int) (*net.Conn, error) {
	return s.readRepository.GetConnectionByUserId(userId)
}
