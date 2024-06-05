package handlers

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"net"
	"sync"
	"testing"
	"time"

	mock_handlers "websocket-confee/internal/handlers/mocks"
	"websocket-confee/internal/services/event/publish/publishers"
)

type handlerTest struct {
	ctrl          *gomock.Controller
	mockRedis     *mock_handlers.MockredisServiceInterface
	mockLog       *mock_handlers.MockloggerInterface
	mockWsService *mock_handlers.MockwsServiceInterface
}

func setUpHandlerMocks(t *testing.T) *handlerTest {
	ctrl := gomock.NewController(t)

	return &handlerTest{
		ctrl:          ctrl,
		mockRedis:     mock_handlers.NewMockredisServiceInterface(ctrl),
		mockLog:       mock_handlers.NewMockloggerInterface(ctrl),
		mockWsService: mock_handlers.NewMockwsServiceInterface(ctrl),
	}
}

func TestSuccessHandle(t *testing.T) {
	mocks := setUpHandlerMocks(t)

	server, _ := net.Listen("tcp", "localhost:8080")
	go func() {
		server.Accept()
	}()

	messageReadEvent := &publishers.MessageReadEvent{
		Event: "MessageRead",
		Data: publishers.MessageReadData{
			ChatId:    1,
			MessageId: 1,
			SessionId: "asdasd",
		},
	}
	msg, _ := json.Marshal(messageReadEvent)
	clientConn, _ := net.Dial("tcp", "localhost:8080")

	mocks.mockWsService.EXPECT().NewReader(clientConn).Times(1)
	mocks.mockWsService.EXPECT().ReadClientMessage(gomock.Any()).Return(msg, nil).AnyTimes()
	mocks.mockLog.EXPECT().Error(gomock.Any()).Times(0)
	mocks.mockRedis.EXPECT().Publish(msg).AnyTimes()

	NewHandler(
		mocks.mockRedis,
		mocks.mockWsService,
		mocks.mockLog,
		make(map[int]net.Conn, 1),
		make(chan struct{}, 1),
		&sync.WaitGroup{},
		clientConn,
		&sync.RWMutex{},
		context.Background(),
	).Handle()

	time.Sleep(20 * time.Nanosecond)
}
