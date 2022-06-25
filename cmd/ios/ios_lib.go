package tun2Simple

import (
	"fmt"
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
			fmt.Println("======>>> release memory for ios")
			debug.FreeOSMemory()
		}
	}()
}

type DeviceI interface {
	stack.DeviceI
}

var _inst *stack.Agent = nil

func InitApp(dev DeviceI) error {
	i, err := stack.SetupAgent(dev)
	if err != nil {
		return err
	}
	_inst = i
	return nil
}

func InputDevData(data []byte) (int, error) {
	fmt.Println("======>>> input from dev to stack", len(data))
	return _inst.ReceiveDevData(data)
}
