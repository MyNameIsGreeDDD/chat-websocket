package handlers

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	mock_handlers "websocket-confee/internal/handlers/mocks"
	"websocket-confee/internal/services/event/publish/publishers"
)

type handlerTest struct {
	ctrl            *gomock.Controller
	mockRedis       *mock_handlers.MockredisServiceInterface
	mockLog         *mock_handlers.MockloggerInterface
	mockWsService   *mock_handlers.MockwsServiceInterface
	mockAuthHandler *mock_handlers.MockauthHandlerInterface
}

func setUpHandlerMocks(t *testing.T) *handlerTest {
	ctrl := gomock.NewController(t)

	return &handlerTest{
		ctrl:            ctrl,
		mockRedis:       mock_handlers.NewMockredisServiceInterface(ctrl),
		mockLog:         mock_handlers.NewMockloggerInterface(ctrl),
		mockWsService:   mock_handlers.NewMockwsServiceInterface(ctrl),
		mockAuthHandler: mock_handlers.NewMockauthHandlerInterface(ctrl),
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
	mocks.mockAuthHandler.EXPECT().Handle(gomock.Any(), gomock.Any()).Times(0)

	NewHandler(
		mocks.mockRedis,
		mocks.mockWsService,
		mocks.mockLog,
		make(map[int]net.Conn, 1),
		make(chan struct{}, 1),
		&sync.WaitGroup{},
		&sync.RWMutex{},
		context.Background(),
		mocks.mockAuthHandler,
	).Handle(clientConn)

	time.Sleep(40 * time.Nanosecond)
}

//func TestAuth_Run(t *testing.T) {
//	mocks := setUpAuthMocks(t)
//
//	token := "asdas"
//	userId := 1
//
//	mockStringCmd := redis.StringCmd{}
//	mockStringCmd.SetVal(strconv.Itoa(userId))
//
//	mocks.mockRedis.EXPECT().HGet(token, "user_id").Return(&mockStringCmd).Times(1)
//	mocks.mockLog.EXPECT().Error(gomock.Any()).Times(0)
//
//	event := AuthEvent{
//		Event: models.Auth,
//		Data: AuthData{
//			Token: token,
//		},
//	}
//	msg, _ := json.Marshal(event)
//
//	_, clientConn := net.Pipe()
//	connections := make(map[int]net.Conn, 1)
//
//	expectedMsg, _ := json.Marshal(models.SuccessResponse{
//		Message: "auth success",
//		Code:    200,
//	})
//
//	mocks.mockWsService.EXPECT().WriteServerBinary(expectedMsg, clientConn).Times(1).Return(nil)
//
//	err := NewAuthHandler(mocks.mockRedis, mocks.mockWsService, connections, &sync.RWMutex{}).Handle(clientConn, msg)
//
//	assert.NoError(t, err)
//	assert.Equal(t, connections[userId], clientConn)
//}
