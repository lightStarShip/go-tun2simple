package stack

import (
	"net"
	"sync"
	"syscall"
	"time"
)

var (
	_once sync.Once
	_inst SimpleStack
)

type ConnProtector func(fd uintptr)
type TunDev interface {
	WriteToTun(p []byte) (n int, err error)
	TunClosed() error
	Protect(fd int32) bool
}
type Wallet interface {
	Address() string
	AesKey() []byte
	MinerNetAddr() string
}

type SimpleStack interface {
	SetupStack(dev TunDev, w Wallet, dnsRule string) error
	WriteToStack(p []byte) (n int, err error)
}

func Inst() SimpleStack {
	_once.Do(func() {
		_inst = newStackV1()
	})
	return _inst
}

func SafeConn(network, rAddr string, connSaver ConnProtector, timeOut time.Duration) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: timeOut,
		Control: func(network, address string, c syscall.RawConn) error {
			if connSaver != nil {
				return c.Control(connSaver)
			}
			return nil
		},
	}
	return d.Dial(network, rAddr)
}
