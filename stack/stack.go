package stack

import (
	"github.com/lightStarShip/go-tun2simple/core"
)

type Agent struct {
	lwipStack core.LWIPStack
}

type DeviceI interface {
	DevToStack(p []byte) (n int, err error)
	StackTODev(data []byte) (int, error)
}

func SetupAgent(di DeviceI) (*Agent, error) {
	lwipStack := core.Inst()
	a := &Agent{
		lwipStack: lwipStack,
	}

	return a, nil
}

func (a *Agent) ReceiveDevData(data []byte) (int, error) {
	return a.lwipStack.InputIpPackets(data)
}
