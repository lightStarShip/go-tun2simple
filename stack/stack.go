package stack

import (
	"github.com/lightStarShip/go-tun2simple/core"
)

type Agent struct {
	devI      DeviceI
	lwipStack core.LWIPStack
}

type DeviceI interface {
	Stack2Dev(data []byte)
}

func SetupAgent(di DeviceI) (*Agent, error) {
	lwipStack := core.Inst()
	a := &Agent{
		devI:      di,
		lwipStack: lwipStack,
	}
	go a.monitorOutput()
	return a, nil
}

func (a *Agent) ReceiveDevData(data []byte) (int, error) {
	return a.lwipStack.InputIpPackets(data)
}

func (a *Agent) monitorOutput() {
	data := a.lwipStack.ReadOutIpPackets()
	a.devI.Stack2Dev(data)
}
