package stack

import (
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"github.com/redeslab/go-simple/account"
	"net"
)

func newStackV1() SimpleStack {

	s := &stackV1{
		lwipStack: core.Inst(),
	}
	return s
}

type stackV1 struct {
	lwipStack core.LWIPStack
	connSaver ConnProtector
	selfId    account.ID
	aesKey    []byte
	minerAddr string
}

func (s1 *stackV1) SetupStack(dev TunDev, w Wallet, rules string) error {
	core.RegisterOutputFn(func(data []byte) (int, error) {
		return dev.WriteToTun(data)
	})

	s1.connSaver = func(fd uintptr) {
		dev.Protect(int32(fd))
	}
	s1.selfId = account.ID(w.Address())
	s1.aesKey = w.AesKey()
	s1.minerAddr = w.MinerNetAddr()

	utils.LogInst().Debugf("======>>> stack param: sid:%s mid:%s", s1.selfId, s1.minerAddr)

	core.RegisterTCPConnHandler(s1)

	dns, err := newDnsHandler(s1.connSaver)
	if err != nil {
		return err
	}
	core.RegisterUDPConnHandler(dns)
	RInst().Setup(rules)
	return nil
}

func (s1 *stackV1) WriteToStack(p []byte) (n int, err error) {
	return s1.lwipStack.Write(p)
}

func (s1 *stackV1) Handle(conn net.Conn, target *net.TCPAddr) error {
	matched := RInst().NeedProxy(target.IP.String())
	if len(matched) > 0 {
		return s1.SimpleRelay(conn, target, matched)
	}
	return s1.directRelay(conn, target)
}
