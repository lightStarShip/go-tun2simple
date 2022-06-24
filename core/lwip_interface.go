package core

import (
	"net"
	"sync"
)

const CHECK_TIMEOUTS_INTERVAL = 250 // in millisecond
const TCP_POLL_INTERVAL = 8         // poll every 4 seconds

type LWIPStack interface {
	InputIpPackets([]byte) (int, error)
	Close() error
	RestartTimeouts()
	Accept() (net.Conn, error)
}

var (
	stackInst *lwipStack = nil
	once      sync.Once
)

func Inst() LWIPStack {
	once.Do(func() {
		stackInst = newLWIPStack()
	})
	return stackInst
}
