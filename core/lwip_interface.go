package core

import "C"
import (
	"net"
	"sync"
)

const CHECK_TIMEOUTS_INTERVAL = 250 // in millisecond
const TCP_POLL_INTERVAL = 8         // poll every 4 seconds

func init() {
	// Initialize lwIP.
	//
	// There is a little trick here, a loop interface (127.0.0.1)
	// is created in the initialization stage due to the option
	// `#define LWIP_HAVE_LOOPIF 1` in `lwipopts.h`, so we need
	// not create our own interface.
	//
	// Now the loop interface is just the first element in
	// `C.netif_list`, i.e. `*C.netif_list`.
	lwipInit()

	// Set MTU.
	C.netif_list.mtu = 1500
}

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
