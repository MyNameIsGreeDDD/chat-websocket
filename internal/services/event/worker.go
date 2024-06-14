package event

import (
	"net"
)

type (
	Publisher interface {
		Publish(msg []byte) error
	}
	Receiver interface {
		Receive(msg []byte, conn net.Conn) error
	}
)
