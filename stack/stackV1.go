package stack

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"github.com/redeslab/go-simple/account"
	"net"
)

func newStackV1() SimpleStack {
	return &stackV1{}
}

type stackV1 struct {
	lwipStack core.LWIPStack
	connSaver ConnProtector
	selfId    account.ID
	aesKey    []byte
	minerAddr string
	mtu       int
	ctx       context.Context
	cancel    context.CancelFunc
}

func (s1 *stackV1) SetupStack(dev TunDev, w Wallet) error {
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
		utils.LogInst().Errorf("======>>> stack param invalid aes key:%s==>err:%s", aesStr, err)
		return err
	}
	s1.aesKey = key
	s1.minerAddr = w.MinerNetAddr()
	s1.mtu = dev.MTU()
	if s1.mtu < MinMtuVal {
		return fmt.Errorf("======>>> too small mtu")
	}

	utils.LogInst().Infof("======>>> stack param: sid:%s mid:%s mtu:%d", s1.selfId, s1.minerAddr, s1.mtu)

	s1.lwipStack = core.NewStack()

	core.RegisterTCPConnHandler(s1)

	ctx, c := context.WithCancel(context.Background())
	s1.ctx = ctx
	s1.cancel = c
	dns, err := newUdpHandler(s1.connSaver, s1.ctx)
	if err != nil {
		return err
	}
	core.RegisterUDPConnHandler(dns)

	rules := dev.LoadRule()
	RInst().Setup(rules)

	inners := dev.LoadInnerIps()
	IPRuleInst().LoadInners(inners)

	mustHits := dev.LoadMustHitIps()
	IPRuleInst().LoadMustHits(mustHits)

	return nil
}

func (s1 *stackV1) DestroyStack() {
	if s1.cancel != nil {
		s1.cancel()
	}

	if s1.lwipStack != nil {
		s1.lwipStack.Close()
	}

	RInst().Close()
}

func (s1 *stackV1) WriteToStack(p []byte) (n int, err error) {
	return s1.lwipStack.Write(p)
}

func (s1 *stackV1) Handle(conn net.Conn, target *net.TCPAddr) error {
	dnsMatched := RInst().NeedProxy(target.IP.String())
	isMustProxy := IPRuleInst().IsMustHits(target.IP)

	var targetMatchedNetAddr = ""
	var matched = false
	if len(dnsMatched) > 0 {
		matched = true
		targetMatchedNetAddr = fmt.Sprintf("%s:%d", dnsMatched, target.Port)
	} else if isMustProxy == true {
		matched = true
		targetMatchedNetAddr = target.String()
	} else {
		isInner := IPRuleInst().IsInnerIP(target.IP)
		if isInner == false {
			matched = true
			targetMatchedNetAddr = target.String()
		}
	}

	if matched {
		utils.LogInst().Infof("======>>> target is matched target:[%s=>%s]", target.String(), targetMatchedNetAddr)

		tarConn, err := s1.setupSimpleConn(targetMatchedNetAddr)
		if err != nil {
			_ = conn.Close()
			utils.LogInst().Errorf("======>>>proxy sync target[%s=>%s] err:%v", target.String(), targetMatchedNetAddr, err)
			return err
		}

		//go s1.upStream(true, conn, tarConn)
		//go s1.downStream(true, conn, tarConn)

		go s1.relay(conn, tarConn)
		go s1.relay(tarConn, conn)

		return nil
	}

	targetConn, err := SafeConn("tcp", target.String(), s1.connSaver, DialTimeOut)
	if err != nil {
		_ = conn.Close()
		utils.LogInst().Errorf("======>>>tcp dial[%s] err:%v", target.String(), err)
		return err
	}
	utils.LogInst().Infof("======>>> direct relay for target:%s", target.String())
	//
	//go s1.upStream(false, conn, targetConn)
	//go s1.downStream(false, conn, targetConn)

	go s1.relay(conn, targetConn)
	go s1.relay(targetConn, conn)
	return nil
}
