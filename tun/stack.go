package tun

import (
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

type StackAgent struct {
	lwipStack core.LWIPStack
}
type DeviceI interface {
	ReadIPPacketsFromStack(p []byte) (n int, err error)
	WriteIPPacketsToStack(data []byte) (int, error)
}

func SetupTun(di DeviceI) {

}
