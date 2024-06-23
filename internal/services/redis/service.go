package redis

import (
	"encoding/json"
	"errors"
	"os"

	"websocket-confee/internal/adapters"
)

var PubEventChannel = os.Getenv("PUB_EVENTS_REDIS_CHANNEL")
var SubEventChannel = os.Getenv("SUB_EVENTS_REDIS_CHANNEL")

type writeRepositoryInterface interface {
	Publish(message []byte, channel string) error
	StoreSessionId(sessionId string, userId int) error
}

type readRepositoryInterface interface {
	Subscribe(channel ...string) adapters.PubSubInterface
	HGet(key, field string) adapters.StringCmdInterface
	GetUserIdBySessionId(sessionId string) (int, error)
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

func (s *Service) StoreSessionId(sessionId string, userId int) error {
	err := s.writeRepository.StoreSessionId(sessionId, userId)
	if err != nil {
		return errors.New("cant store session id in redis service with err: " + err.Error())
	}

	return nil
}
func (s *Service) GetUserIdBySessionId(sessionId string) (int, error) {
	userId, err := s.readRepository.GetUserIdBySessionId(sessionId)
	if err != nil {
		return 0, errors.New("cant gets user_id by session_id: " + sessionId)
	}
	return userId, nil
}
