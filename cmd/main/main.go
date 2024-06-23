package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/gobwas/ws"
	r_client "github.com/redis/go-redis/v9"

	"websocket-confee/internal/handlers"
	"websocket-confee/internal/models"
	"websocket-confee/internal/repositories/redis/read"
	"websocket-confee/internal/repositories/redis/write"
	"websocket-confee/internal/services/event"
	"websocket-confee/internal/services/event/publish/publishers"
	"websocket-confee/internal/services/event/receive"
	"websocket-confee/internal/services/event/receive/receivers"
	"websocket-confee/internal/services/logger"
	r_service "websocket-confee/internal/services/redis"
	"websocket-confee/internal/services/websocket"
)

const CountConnections = 2500

func main() {
	wsService := websocket.NewWebSocketService()
	log := logger.NewLogger()

	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		panic("cant start redis service")
	}
	client := r_client.NewClient(&r_client.Options{
		Addr: os.Getenv("REDIS_URL") + ":" + os.Getenv("REDIS_HOST"),
		DB:   db,
	})

	redisReadRepository := read.NewRedisReadRepository(client)
	redisWriteRepository := write.NewRedisWriteRepository(client)

	redisService := r_service.NewRedisService(log, redisWriteRepository, redisReadRepository)

	subsWG := sync.WaitGroup{}
	subsPool := make(chan struct{}, 200)

	globalCtx, globalCancel := context.WithCancel(context.Background())
	listenerCtx, _ := context.WithCancel(globalCtx)

	allowedReceivers := initReceivers(wsService)
	allowedPublishers := initPublishers(redisService)
	connections := make(map[int]net.Conn, CountConnections)

	receive.RegisterEventListener(redisService, wsService, log, connections, subsPool, &subsWG, listenerCtx, allowedReceivers).Run()

	ln, err := net.Listen("tcp", "localhost:8080")
	fmt.Println("Start listen on :8080")

	if err != nil {
		panic("failed listen")
	}

	if os.Getenv("APP_ENV") == "dev" {
		go func() {
			fmt.Println("Starting pprof on :6060")
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				fmt.Println("cant start pprof")
			}
		}()
	}

	u := &ws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	connWg := sync.WaitGroup{}
	connsPool := make(chan struct{}, CountConnections)

	initGraceFullShutDown(&connWg, &subsWG, ln, connections, log, wsService, globalCancel)

	connMutex := &sync.RWMutex{}
	publisherCtx, _ := context.WithCancel(globalCtx)

	eventHandler := handlers.NewHandler(
		redisService,
		wsService,
		log,
		connections,
		connsPool,
		&connWg,
		connMutex,
		publisherCtx,
		allowedPublishers,
	)

	for {
		conn, err := wsService.AcceptConnection(ln, u)
		if err != nil {
			log.Error(fmt.Sprintf("error while accept connection: %s", err))
			continue
		}

		eventHandler.Handle(conn)
	}

	connWg.Wait()
	subsWG.Wait()
}

func initReceivers(wsService *websocket.Service) map[string]event.Receiver {
	allowedReceivers := make(map[string]event.Receiver, 1)
	allowedReceivers[models.MessageRead] = receivers.NewMessageReadReceiver(wsService)

	return allowedReceivers
}

func initPublishers(service *r_service.Service) map[string]event.Publisher {
	allowedPublishers := make(map[string]event.Publisher, 1)
	allowedPublishers[models.MessageRead] = publishers.NewMessageReadPublisher(service)

	return allowedPublishers
}

func initGraceFullShutDown(
	consWg, subsWg *sync.WaitGroup,
	tcpConn net.Listener,
	connections map[int]net.Conn,
	logger *logger.Logger,
	wsService *websocket.Service,
	globalCancel context.CancelFunc,
) {
	shutdownChan := make(chan os.Signal, 2)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		fmt.Println("Программа получила сигнал остановки. Начало процедуры graceful shutdown.")

		globalCancel()

		go logger.BatchSave()

		connWG := sync.WaitGroup{}
		connPool := make(chan struct{}, 1000)

		for _, conn := range connections {
			connWG.Add(1)
			connPool <- struct{}{}
			go func() {
				defer func() { <-connPool }()
				defer connWG.Done()

				wsService.WriteServerBinary([]byte("connection closed"), conn)
			}()
		}

		connWG.Wait()
		consWg.Wait()
		subsWg.Wait()
		tcpConn.Close()

		fmt.Println("Graceful shutdown окончен")
		os.Exit(0)
	}()
}
