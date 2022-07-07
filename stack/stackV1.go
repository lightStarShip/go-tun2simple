package stack

import (
	"encoding/hex"
	"fmt"
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
		dev.SafeConn(int32(fd))
	}
	s1.selfId = account.ID(w.Address())
	aesStr := w.AesKeyBase64()
	key, err := hex.DecodeString(aesStr)
	if err != nil {
		utils.LogInst().Errorf("======>>> stack param invalid aes key:%s", aesStr)
		return err
	}
	s1.aesKey = key
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
		nameTarget := fmt.Sprintf("%s:%d", matched, target.Port)
		tarConn, err := s1.setupSimpleConn(nameTarget)
		if err != nil {
			_ = conn.Close()
			utils.LogInst().Errorf("======>>>proxy sync target[%s=>%s] err:%v", target.String(), nameTarget, err)
			return err
		}
		utils.LogInst().Infof("======>>> proxy for target:[%s=>%s]", target.String(), nameTarget)

		go s1.relay(conn, tarConn)
		return nil
	}

	targetConn, err := SafeConn("tcp", target.String(), s1.connSaver, DialTimeOut)
	if err != nil {
		_ = conn.Close()
		utils.LogInst().Errorf("======>>>tcp dial[%s] err:%v", target.String(), err)
		return err
	}
	utils.LogInst().Infof("======>>> direct relay for target:%s", target.String())

	go s1.relay(conn, targetConn)
	return nil
}
