package receivers

import (
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"websocket-confee/internal/models"
	mock_receive "websocket-confee/internal/services/event/receive/mocks"
)

type receiverTest struct {
	ctrl          *gomock.Controller
	mockWsService *mock_receive.MockwsServiceInterface
}

func setUpMocks(t *testing.T) *receiverTest {
	ctrl := gomock.NewController(t)

	return &receiverTest{
		ctrl:          ctrl,
		mockWsService: mock_receive.NewMockwsServiceInterface(ctrl),
	}
}

func TestSuccessRunMessageRead(t *testing.T) {
	mocks := setUpMocks(t)
	defer mocks.ctrl.Finish()

	userId := 1
	msg := MessageReadEvent{
		Event:  models.MessageRead,
		UserId: userId,
		Data: MessageReadData{
			InfoForClient: InfoForClient{
				ChatId:     1,
				MessageId:  1,
				ReadUserId: userId,
			},
		},
		Socket: "sdfsdf",
	}
	marshalMsg, _ := json.Marshal(msg)

	clientConn, _ := net.Dial("tcp", "localhost:8080")

	mocks.mockWsService.EXPECT().WriteClientBinary(marshalMsg, clientConn).Times(1).Return(nil)
	err := NewMessageReadReceiver(marshalMsg, clientConn, mocks.mockWsService).Run()
	assert.NoError(t, err)
}
