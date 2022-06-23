package tun2Simple

import (
	"github.com/lightStarShip/go-tun2simple/stack"
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
			debug.FreeOSMemory()
		}
	}()
}

var _inst *stack.Agent = nil

func InitApp(dev stack.DeviceI) error {
	i, err := stack.SetupAgent(dev)
	if err != nil {
		return err
	}
	_inst = i
	return nil
}
