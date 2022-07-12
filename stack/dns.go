package stack

import (
	"encoding/hex"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"sync"
)

const (
	COMMON_DNS_PORT = 53
)

type dnsHandler struct {
	sync.Mutex
	saver       ConnProtector
	pivot       *net.UDPConn
	cache       map[uint16]core.UDPConn
	redirectMap map[string]net.Conn
}

func newDnsHandler(saver ConnProtector) (core.UDPConnHandler, error) {
	bindAddr := &net.UDPAddr{IP: nil, Port: 0}
	pc, err := net.ListenUDP("udp4", bindAddr)
	if err != nil {
		utils.LogInst().Errorf("======>>>DNS ListenUDP err:=>%s", err.Error())
		return nil, err
	}
	raw, err := pc.SyscallConn()
	if err != nil {
		utils.LogInst().Errorf("======>>>DNS SyscallConn err:=>%s", err.Error())
		return nil, err
	}
	if err := raw.Control(saver); err != nil {
		utils.LogInst().Errorf("======>>>DNS raw Control err:=>%s", err.Error())
		return nil, err
	}

	handler := &dnsHandler{
		pivot:       pc,
		saver:       saver,
		cache:       make(map[uint16]core.UDPConn),
		redirectMap: make(map[string]net.Conn),
	}
	go handler.waitResponse()
	utils.LogInst().Debugf("======>>> create dns handler[%s] success:=>", pc.LocalAddr().String())
	return handler, nil
}

func udpID(src, dst string) string {
	return fmt.Sprintf("%s->%s", src, dst)
}

func (dh *dnsHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	utils.LogInst().Debugf("======>>>Connect:%s------>>>%s", conn.LocalAddr().String(), target.String())
	if target.Port == COMMON_DNS_PORT {
		return nil
	}
	peerUdp, err := SafeConn("udp", target.String(), dh.saver, DialTimeOut)
	if err != nil {
		return err
	}
	id := udpID(conn.LocalAddr().String(), target.String())
	dh.Lock()
	dh.redirectMap[id] = peerUdp
	dh.Unlock()
	go dh.receiveFromTarget(conn, peerUdp, target)
	return nil
}

func (dh *dnsHandler) close() {
	utils.LogInst().Warnf("======>>>dns handler quit......")
	dh.Lock()
	for _, conn := range dh.cache {
		conn.Close()
	}
	dh.Unlock()
	dh.pivot.Close()
}

func (dh *dnsHandler) waitResponse() {
	utils.LogInst().Infof("======>>> dns wait thread start work......")
	defer utils.LogInst().Infof("======>>> dns wait thread off work......")
	buf := utils.NewBytes(utils.BufSize)

	defer func() {
		dh.close()
		utils.FreeBytes(buf)
	}()

	for {
		n, addr, err := dh.pivot.ReadFromUDP(buf)
		if err != nil {
			utils.LogInst().Errorf("failed to read UDP data from remote: %v", err)
			return
		}
		msg := &dnsmessage.Message{}
		if err := msg.Unpack(buf[:n]); err != nil {
			utils.LogInst().Errorf("======>>>Unpack dns response err:%s\n%s\n", err.Error(), hex.EncodeToString(buf[:n]))
			continue
		}
		dh.Lock()
		conn, ok := dh.cache[msg.ID]
		if !ok {
			dh.Unlock()
			utils.LogInst().Warnf("======>>> no such[%d] cache item", msg.ID)
			continue
		}
		delete(dh.cache, msg.ID)
		dh.Unlock()

		_, err = conn.WriteFrom(buf[:n], addr)
		if err != nil {
			utils.LogInst().Errorf("======>>>dns proxy write back err:%s", err.Error())
			continue
		}
		RInst().ParseDns(msg)
	}
}

func (dh *dnsHandler) receiveFromTarget(conn core.UDPConn, peerUdp net.Conn, target *net.UDPAddr) {
	buf := utils.NewBytes(utils.BufSize)
	defer utils.FreeBytes(buf)
	utils.LogInst().Warnf("======>>>prepare to read udp for target:=>%s", target.String())

	defer dh.clearUdpRelay(target.String())
	defer conn.Close()
	for {
		n, err := peerUdp.Read(buf)
		if err != nil {
			utils.LogInst().Warnf("======>>>udp relay app<------target err:=>%s", err.Error())
			return
		}
		_, err = conn.WriteFrom(buf[:n], target)
		if err != nil {
			utils.LogInst().Warnf("======>>>udp relay app<------target err:=>%s", err.Error())
			return
		}
	}
}

func (dh *dnsHandler) clearUdpRelay(target string) {
	peerUdp, ok := dh.redirectMap[target]
	if !ok {
		return
	}

	peerUdp.Close()
	delete(dh.redirectMap, target)
}

func (dh *dnsHandler) forwardToTarget(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	id := udpID(conn.LocalAddr().String(), addr.String())
	dh.Lock()
	peerUdp, ok := dh.redirectMap[id]
	if !ok {
		dh.Unlock()
		conn.Close()
		err := fmt.Errorf("no peer udp relay found for addr:%s", id)
		utils.LogInst().Warnf("======>>>udp relay app------>target err:=>%s", err.Error())
		return err
	}
	dh.Unlock()
	_, err := peerUdp.Write(data)
	if err == nil {
		return nil
	}
	dh.clearUdpRelay(addr.String())
	conn.Close()
	utils.LogInst().Warnf("======>>>udp relay app------>target[id=%s] peer write err:=>%s", id, err.Error())
	return err
}

func (dh *dnsHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	utils.LogInst().Debugf("======>>>ReceiveTo %s------>>>%s", conn.LocalAddr().String(), addr)

	if addr.Port != COMMON_DNS_PORT {
		return dh.forwardToTarget(conn, data, addr)
	}

	_, err := dh.pivot.WriteToUDP(data, addr)
	if err != nil {
		utils.LogInst().Errorf("======>>>dns forward err:%s\n%s\n", err.Error(), hex.EncodeToString(data))
		return err
	}

	msg := &dnsmessage.Message{}
	if err := msg.Unpack(data); err != nil {
		utils.LogInst().Errorf("======>>>Unpack dns request err:%s", err.Error(), hex.EncodeToString(data))
		return err
	}
	utils.LogInst().Debugf("======>>>dns[%d] questions:%s =>", msg.ID, msg.Questions)

	dh.Lock()
	dh.cache[msg.ID] = conn
	dh.Unlock()

	return nil
}
