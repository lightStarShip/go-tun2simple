package core

import (
	"fmt"
	"net"
	"sync"
)

const (
	CHECK_TIMEOUTS_INTERVAL = 250 // in millisecond
	TCP_POLL_INTERVAL       = 8   // poll every 4 seconds
)

type DebugLog func(isOpen bool, a ...any)

type LWIPStack interface {
	InputIpPackets([]byte) (int, error)
	Close() error
	RestartTimeouts()
	Accept() (net.Conn, error)
	OutputIpPackets() []byte
}

var (
	stackInst   *lwipStack = nil
	once        sync.Once
	_console    DebugLog = nil
	detailDebug          = true
)

func Inst() LWIPStack {
	once.Do(func() {
		_console = defaultLog
		stackInst = newLWIPStack()
	})
	return stackInst
}

func defaultLog(isOpen bool, a ...any) {
	if !isOpen {
		return
	}
	fmt.Println(a...)
}

func RegLogFunc(dl DebugLog) {
	_console = dl
}
