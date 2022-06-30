package tun2Simple

import (
	"errors"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/proxy/redirect"
	"github.com/lightStarShip/go-tun2simple/utils"
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

const (
	COMMON_DNS_PORT = 53
	dnsHeaderLength = 12
	dnsMaskQr       = uint8(0x80)
	dnsMaskTc       = uint8(0x02)
	dnsMaskRcode    = uint8(0x0F)
)

var _iosApp *iosApp = nil

type Tunnel interface {
	Write(data []byte) (int, error)
}

type iosApp struct {
	lwipStack core.LWIPStack
	dev       TunnelDev
}
type TunnelDev interface {
	io.WriteCloser
	Log(s string)
}

func console(msg string, a ...any) {
	log := fmt.Sprintf(msg, a...)
	_iosApp.dev.Log(log)
}

func NewTunnel(tunWriter TunnelDev) (Tunnel, error) {
	if tunWriter == nil {
		return nil, errors.New("Must provide a TUN writer")
	}

	core.RegisterOutputFn(func(data []byte) (int, error) {
		//console("======>>>data raw:", hex.EncodeToString(data))
		return tunWriter.Write(data)
	})
	lwipStack := core.Inst()
	t := &iosApp{
		lwipStack,
		tunWriter}
	core.RegisterTCPConnHandler(t)
	core.RegisterUDPConnHandler(redirect.NewUDPHandler(10))
	_iosApp = t

	utils.LogInst().InitParam(utils.DEBUG, console)

	return t, nil
}

func (t *iosApp) Write(data []byte) (int, error) {
	return t.lwipStack.Write(data)
}

func (t *iosApp) Handle(conn net.Conn, target *net.TCPAddr) error {
	utils.LogInst().Debugf("======>>>Handle implement me:", conn.LocalAddr(), target.String())
	return nil
}
