package receive

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"
	"websocket-confee/internal/services/event"

	"github.com/golang/mock/gomock"
	r_client "github.com/redis/go-redis/v9"
	_ "github.com/stretchr/testify/assert"

	mock_adapters "websocket-confee/internal/adapters/mocks"
	"websocket-confee/internal/models"
	mock_receive "websocket-confee/internal/services/event/receive/mocks"
	"websocket-confee/internal/services/event/receive/receivers"
)

type receiverTest struct {
	ctrl          *gomock.Controller
	mockRedis     *mock_receive.MockredisServiceInterface
	mockLog       *mock_receive.MockloggerInterface
	mockPubSub    *mock_adapters.MockPubSubInterface
	mockWsService *mock_receive.MockwsServiceInterface
}

func setUpMocks(t *testing.T) *receiverTest {
	ctrl := gomock.NewController(t)

	return &receiverTest{
		ctrl:          ctrl,
		mockRedis:     mock_receive.NewMockredisServiceInterface(ctrl),
		mockLog:       mock_receive.NewMockloggerInterface(ctrl),
		mockPubSub:    mock_adapters.NewMockPubSubInterface(ctrl),
		mockWsService: mock_receive.NewMockwsServiceInterface(ctrl),
	}
}

func TestSuccessRun(t *testing.T) {
	mocks := setUpMocks(t)
	defer mocks.ctrl.Finish()

	userId := 1
	connections := map[int]net.Conn{
		userId: nil,
	}

	server, _ := net.Listen("tcp", "localhost:8080")
	go func() {
		serverConn, _ := server.Accept()
		connections[userId] = serverConn
	}()

	msgChan := make(chan *r_client.Message, 1)
	msg := receivers.MessageReadEvent{
		Event:  models.MessageRead,
		UserId: userId,
		Data: receivers.MessageReadData{
			InfoForClient: receivers.InfoForClient{
				ChatId:     1,
				MessageId:  1,
				ReadUserId: userId,
			},
		},
		Socket: "sdfsdf",
	}
	marshalMsg, _ := json.Marshal(msg)
	msgChan <- &r_client.Message{Payload: string(marshalMsg)}

	mockPubSub := mock_adapters.NewMockPubSubInterface(mocks.ctrl)
	mockPubSub.EXPECT().Channel().Return(msgChan)

	net.Dial("tcp", "localhost:8080")

	mocks.mockRedis.EXPECT().Subscribe(gomock.Any()).Return(mockPubSub)
	mocks.mockLog.EXPECT().Error(gomock.Any()).Times(0)
	mocks.mockWsService.EXPECT().WriteServerBinary(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	eventSub := RegisterEventListener(
		mocks.mockRedis,
		mocks.mockWsService,
		mocks.mockLog,
		connections,
		make(chan struct{}, 5),
		&sync.WaitGroup{},
		context.Background(),
		map[string]event.Receiver{},
	)

	eventSub.Run()

	time.Sleep(time.Millisecond * 1)
}
