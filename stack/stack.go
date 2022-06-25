package stack

import (
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"net"
)

type Agent struct {
	devI      DeviceI
	lwipStack core.LWIPStack
	sig       chan struct{}
}

type DeviceI interface {
	Stack2Dev(data []byte)
	StackClosed()
	Log(s string)
}

func SetupAgent(di DeviceI) (*Agent, error) {
	lwipStack := core.Inst()
	a := &Agent{
		devI:      di,
		lwipStack: lwipStack,
		sig:       make(chan struct{}, 1),
	}
	core.RegLogFunc(a.cliLog)
	go a.monitorOutput()
	go a.listening()
	return a, nil
}

func (a *Agent) ReceiveDevData(data []byte) (int, error) {
	return a.lwipStack.InputIpPackets(data)
}

func (a *Agent) cliLog(isOpen bool, args ...any) {
	if isOpen {
		return
	}
	a.devI.Log(fmt.Sprintln(args...))
}

func (a *Agent) monitorOutput() {

	for {
		select {
		case <-a.sig:
			return
		default:
			data := a.lwipStack.OutputIpPackets()
			if data == nil {
				a.finished()
				return
			}
			a.devI.Stack2Dev(data)
		}
	}
}

func (a *Agent) finished() {

	if a.lwipStack == nil {
		return
	}
	a.lwipStack.Close()
	a.devI.StackClosed()
	close(a.sig)
	a.sig = nil
	a.lwipStack = nil
	a.devI = nil
}

func (a *Agent) listening() {

	for {
		select {
		case <-a.sig:
			return

		default:
			conn, err := a.lwipStack.Accept()
			if err != nil {
				a.finished()
				return
			}
			a.relay(conn)
		}
	}
}

func (a *Agent) relay(conn net.Conn) {
	a.cliLog(true, "======>>>new conn:", conn.LocalAddr().String(), conn.RemoteAddr().String())
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	a.cliLog(true, "======>>>new conn:", buf[:n], n, err)
}
