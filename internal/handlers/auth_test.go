package handlers

import (
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	mock_adapters "websocket-confee/internal/adapters/mocks"
	mock_handlers "websocket-confee/internal/handlers/mocks"
	"websocket-confee/internal/models"
)

type authTest struct {
	ctrl          *gomock.Controller
	mockRedis     *mock_handlers.MockredisServiceInterface
	mockLog       *mock_handlers.MockloggerInterface
	mockStringCmd *mock_adapters.MockStringCmdInterface
	mockWsService *mock_handlers.MockwsServiceInterface
}

func setUpAuthMocks(t *testing.T) *authTest {
	ctrl := gomock.NewController(t)

	return &authTest{
		ctrl:          ctrl,
		mockRedis:     mock_handlers.NewMockredisServiceInterface(ctrl),
		mockLog:       mock_handlers.NewMockloggerInterface(ctrl),
		mockStringCmd: mock_adapters.NewMockStringCmdInterface(ctrl),
		mockWsService: mock_handlers.NewMockwsServiceInterface(ctrl),
	}
}

func TestAuth_Run(t *testing.T) {
	mocks := setUpAuthMocks(t)

	token := "asdas"
	userId := 1

	mockStringCmd := redis.StringCmd{}
	mockStringCmd.SetVal(strconv.Itoa(userId))

	mocks.mockRedis.EXPECT().HGet(token, "user_id").Return(&mockStringCmd).Times(1)
	mocks.mockLog.EXPECT().Error(gomock.Any()).Times(0)

	event := AuthEvent{
		Event: models.Auth,
		Data: AuthData{
			Token: token,
		},
	}
	msg, _ := json.Marshal(event)

	_, clientConn := net.Pipe()
	connections := make(map[int]net.Conn, 1)

	expectedMsg, _ := json.Marshal(models.SuccessResponse{
		Message: "auth success",
		Code:    200,
	})

	mocks.mockWsService.EXPECT().WriteClientBinary(expectedMsg, clientConn).Times(1).Return(nil)

	NewAuthHandler(mocks.mockRedis, mocks.mockWsService, connections, &sync.RWMutex{}, clientConn, msg).Handle()
	assert.Equal(t, connections[userId], clientConn)
}
