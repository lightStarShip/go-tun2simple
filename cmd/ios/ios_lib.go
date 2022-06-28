package tun2Simple

import (
	"errors"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"io"
	"net"
	"runtime/debug"
	"time"
)

func init() {
	// Apple VPN extensions have a memory limit of 15MB. Conserve memory by increasing garbage
	// collection frequency and returning memory to the OS every minute.
	debug.SetGCPercent(10)
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for range ticker.C {
			fmt.Println("======>>> release memory for ios")
			debug.FreeOSMemory()
		}
	}()
}

type Tunnel interface {
	Write(data []byte) (int, error)
}

type outlinetunnel struct {
	lwipStack core.LWIPStack
	dev       TunnelDev
}
type TunnelDev interface {
	io.WriteCloser
	Log(s string)
}

func NewTunnel(tunWriter TunnelDev) (Tunnel, error) {
	if tunWriter == nil {
		return nil, errors.New("Must provide a TUN writer")
	}

	core.RegisterOutputFn(func(data []byte) (int, error) {
		return tunWriter.Write(data)
	})
	lwipStack := core.NewLWIPStack()
	t := &outlinetunnel{
		lwipStack,
		tunWriter}
	t.registerConnectionHandlers()
	return t, nil
}

func (t *outlinetunnel) Write(data []byte) (int, error) {
	return t.lwipStack.Write(data)
}

// Registers UDP and TCP Shadowsocks connection handlers to the tunnel's host and port.
// Registers a DNS/TCP fallback UDP handler when UDP is disabled.
func (t *outlinetunnel) registerConnectionHandlers() {
	core.RegisterTCPConnHandler(t)
	core.RegisterUDPConnHandler(t)
}

func (t *outlinetunnel) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	t.dev.Log("======>>>Connect implement me")
	return nil
}

func (t *outlinetunnel) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	t.dev.Log("======>>>ReceiveTo implement me")
	return nil
}

func (t *outlinetunnel) Handle(conn net.Conn, target *net.TCPAddr) error {
	t.dev.Log("======>>>Handle implement me")
	return nil
}
