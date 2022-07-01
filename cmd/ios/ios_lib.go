package tun2Simple

import (
	"errors"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"io"
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
			utils.LogInst().Infof("======>>> release memory for ios")
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
	LoadRule() string
}

func console(msg string, a ...any) {
	log := fmt.Sprintf(msg, a...)
	_iosApp.dev.Log(log)
}

func NewTunnel(dev TunnelDev, logLevel int8) (Tunnel, error) {
	if dev == nil {
		return nil, errors.New("Must provide a TUN writer")
	}
	utils.LogInst().InitParam(utils.LogLevel(logLevel), console)

	core.RegisterOutputFn(func(data []byte) (int, error) {
		//utils.LogInst().Debugf("======>>>RegisterOutputFn:%s", hex.EncodeToString(data))
		return dev.Write(data)
	})
	lwipStack := core.Inst()
	t := &iosApp{
		lwipStack,
		dev}
	core.RegisterTCPConnHandler(newTCPHandler())
	core.RegisterUDPConnHandler(NewDnsHandler())
	_iosApp = t

	rules := dev.LoadRule()
	RInst().Setup(rules)
	return t, nil
}

func (t *iosApp) Write(data []byte) (int, error) {
	return t.lwipStack.Write(data)
}
