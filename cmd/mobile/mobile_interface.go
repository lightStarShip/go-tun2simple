package tun2Simple

import (
	"errors"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/stack"
	"github.com/lightStarShip/go-tun2simple/utils"
	"runtime/debug"
	"time"
)

func init() {
	// Apple VPN extensions have a memory limit of 50MB. Conserve memory by increasing garbage
	// collection frequency and returning memory to the OS every minute.
	debug.SetGCPercent(10)
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for range ticker.C {
			utils.LogInst().Infof("======>>> release memory for garbage collection")
			debug.FreeOSMemory()
		}
	}()
}

type ExtensionI interface {
	stack.TunDev
	stack.Wallet
	Log(s string)
}

func InitEx(exi ExtensionI, logLevel int8) error {
	if exi == nil {
		return errors.New("invalid tun device")
	}
	utils.LogInst().InitParam(utils.LogLevel(logLevel), func(msg string, args ...any) {
		log := fmt.Sprintf(msg, args...)
		exi.Log(log)
	})
	return stack.Inst().SetupStack(exi, exi)
}

func WritePackets(data []byte) (int, error) {
	return stack.Inst().WriteToStack(data)
}
