gen:
	mockgen --source=internal/services/event/receive/receiver.go\
 	--destination=internal/services/event/receive/mocks/receiver.go &&\
 	mockgen --source=internal/adapters/redis.go\
    --destination=internal/adapters/mocks/redis.go &&\
    mockgen --source=internal/handlers/handler.go\
    --destination=internal/handlers/mocks/handler.go &&\
    mockgen --source=internal/services/event/receive/receivers/message_read.go\
    --destination=internal/services/event/receive/receivers/mocks/message_read .go