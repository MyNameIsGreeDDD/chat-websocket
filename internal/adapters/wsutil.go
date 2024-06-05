package adapters

import (
	"github.com/gobwas/ws"
)

type ReaderInterface interface {
	NextFrame() (hdr ws.Header, err error)
	Read(p []byte) (n int, err error)
}
