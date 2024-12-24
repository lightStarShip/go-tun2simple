package stack

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"github.com/redeslab/go-simple/account"
	"net"
)

func newStackV1() SimpleStack {
	return &stackV1{
		counter: make(map[int]int),
	}
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
	counter   map[int]int
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

	utils.LogInst().Infof("======>>> stack param: sid:%s mid:%s mtu:%d", s1.selfId, s1.minerAddr, s1.mtu)

	s1.lwipStack = core.NewStack()

	core.RegisterTCPConnHandler(s1)

	ctx, c := context.WithCancel(context.Background())
	s1.ctx = ctx
	s1.cancel = c

	core.RegisterUDPConnHandler(newUdpRelay(s1.connSaver))

	rules := dev.LoadRule()
	RInst().Setup(rules)

	inners := dev.LoadInnerIps()
	IPRuleInst().LoadInners(inners)

	mustHits := dev.LoadMustHitIps()
	IPRuleInst().LoadMustHits(mustHits)

	return nil
}

func (s1 *stackV1) SetGlobal(global bool) {
	IPRuleInst().SetGlobal(global)
	utils.LogInst().Debugf("======>>> change global status: network layer: %t", global)
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

func (s1 *stackV1) WriteToStack(buf []byte) (n int, err error) {

	var ip4 *layers.IPv4 = nil
	packet := gopacket.NewPacket(buf, layers.LayerTypeIPv4, gopacket.Default)

	if ip4Layer := packet.Layer(layers.LayerTypeIPv4); ip4Layer != nil {
		ip4 = ip4Layer.(*layers.IPv4)
	} else {
		utils.LogInst().Infof("======>>> Unsupported network layer: \n%s\n", packet.Dump())
		return
	}

	var tcp *layers.TCP = nil
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp = tcpLayer.(*layers.TCP)
		srcPort := int(tcp.SrcPort)
		tcpLen := len(tcp.Payload)
		s1.counter[srcPort]++
		if (s1.counter[srcPort] == 3 || s1.counter[srcPort] == 4) && tcpLen > 10 { //
			host := utils.ParseHost(tcp.Payload)
			if len(host) > 0 {
				utils.LogInst().Infof("======>>> Found[%d] host[%s] success for[%d->%s]",
					s1.counter[srcPort], host, srcPort, ip4.DstIP.String())
				RInst().DirectIPAndHost(host+".", ip4.DstIP.String())
			}
		}
	}

	return s1.lwipStack.Write(buf)
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
		utils.LogInst().Infof("======>>>[TCP] target is matched target:[%s=>%s]", target.String(), targetMatchedNetAddr)

		tarConn, err := s1.setupSimpleConn(targetMatchedNetAddr)
		if err != nil {
			_ = conn.Close()
			utils.LogInst().Errorf("======>>>[TCP]proxy sync target[%s=>%s] err:%v", target.String(), targetMatchedNetAddr, err)
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
		utils.LogInst().Errorf("======>>>[TCP] dial[%s] err:%v", target.String(), err)
		return err
	}
	utils.LogInst().Infof("======>>>[TCP] direct relay for target:%s", target.String())
	//
	//go s1.upStream(false, conn, targetConn)
	//go s1.downStream(false, conn, targetConn)

	go s1.relay(conn, targetConn)
	go s1.relay(targetConn, conn)
	return nil
}
