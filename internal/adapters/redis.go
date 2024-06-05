package adapters

import r_client "github.com/redis/go-redis/v9"

type PubSubInterface interface {
	Channel(opts ...r_client.ChannelOption) <-chan *r_client.Message
}

type StringCmdInterface interface {
	Int() (int, error)
}
