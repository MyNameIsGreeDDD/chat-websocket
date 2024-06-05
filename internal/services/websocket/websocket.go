package websocket

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"websocket-confee/internal/adapters"
)

type Service struct {
}

func NewWebSocketService() *Service {
	return &Service{}
}

func (s *Service) AcceptConnection(ln net.Listener, u *ws.Upgrader) (net.Conn, error) {
	conn, err := ln.Accept()
	if err != nil {
		conn.Close()

		return nil, err
	}

	_, err = u.Upgrade(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (s *Service) WriteClientText(msg []byte, conn net.Conn) error {
	return s.writeClientMessage(msg, conn, ws.OpText)
}

func (s *Service) WriteClientClose(msg []byte, conn net.Conn) error {
	return s.writeClientMessage(msg, conn, ws.OpClose)
}

func (s *Service) WriteClientBinary(msg []byte, conn net.Conn) error {
	return s.writeClientMessage(msg, conn, ws.OpBinary)
}

func (s *Service) writeClientMessage(msg []byte, conn net.Conn, opCode ws.OpCode) error {
	hr := ws.Header{
		Fin:    true,
		OpCode: opCode,
		Masked: true,
		Length: int64(len(msg)),
	}

	if err := ws.WriteHeader(conn, hr); err != nil {
		return errors.New("websocket: failed to write header: " + err.Error())
	}

	if _, err := conn.Write(msg); err != nil {
		return errors.New("cant write message: " + err.Error())
	}

	return nil
}

func (s *Service) ReadClientMessage(reader adapters.ReaderInterface) ([]byte, error) {
	header, err := reader.NextFrame()
	if err != nil {
		return nil, err
	}

	msg := make([]byte, header.Length)
	if _, err := io.ReadFull(reader, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *Service) NewReader(conn net.Conn) adapters.ReaderInterface {
	rd := wsutil.NewReader(conn, ws.StateServerSide)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	return rd
}
