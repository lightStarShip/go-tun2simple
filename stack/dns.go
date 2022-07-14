package stack

import (
	"encoding/hex"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"os"
	"sync"
	"time"
)

const (
	COMMON_DNS_PORT = 53
	ExpireTime      = time.Minute * 3
)

type dnsConn struct {
	core.UDPConn
	updateTime time.Time
}
type dnsHandler struct {
	cLocker     sync.RWMutex
	rLocker     sync.RWMutex
	saver       ConnProtector
	pivot       *net.UDPConn
	cache       map[uint16]*dnsConn
	redirectMap map[string]net.Conn
	expire      *time.Ticker
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
		cache:       make(map[uint16]*dnsConn),
		redirectMap: make(map[string]net.Conn),
		expire:      time.NewTicker(ExpireTime),
	}
	go handler.waitResponse()
	go handler.expireConn()
	utils.LogInst().Debugf("======>>> create dns handler[%s] success:=>", pc.LocalAddr().String())
	return handler, nil
}

func udpID(src, dst string) string {
	return fmt.Sprintf("%s->%s", src, dst)
}
func (dh *dnsHandler) expireConn() {
	for {
		select {
		case time := <-dh.expire.C:
			utils.LogInst().Infof("======>>> timer[%s] cleaner start:=>", time.String())
			toDelete := make([]uint16, 0)
			for idx, conn := range dh.cache {
				if time.Sub(conn.updateTime) <= ExpireTime {
					utils.LogInst().Debugf("======>>> dns[%d] still ok:=>", idx)
					continue
				}
				utils.LogInst().Infof("======>>> dns[%d] need to be deleted:=>", idx)
				toDelete = append(toDelete, idx)
				conn.Close()
			}

			if len(toDelete) == 0 {
				continue
			}

			dh.cLocker.Lock()
			for _, idx := range toDelete {
				delete(dh.cache, idx)
			}
			dh.cLocker.Unlock()
		}
	}
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
	dh.rLocker.Lock()
	dh.redirectMap[id] = peerUdp
	dh.rLocker.Unlock()
	go dh.receiveFromTarget(conn, peerUdp, target)
	return nil
}

func (dh *dnsHandler) close() {
	utils.LogInst().Warnf("======>>>dns handler quit......")
	dh.cLocker.Lock()
	for _, conn := range dh.cache {
		conn.Close()
	}
	dh.cLocker.Unlock()
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
			utils.LogInst().Errorf("======>>>failed to read UDP data from remote: %v", err)
			os.Exit(-1)
		}
		msg := &dnsmessage.Message{}
		if err := msg.Unpack(buf[:n]); err != nil {
			utils.LogInst().Errorf("======>>>Unpack dns response err:%s\n%s\n", err.Error(), hex.EncodeToString(buf[:n]))
			continue
		}
		utils.LogInst().Debugf("======>>>dns[%d] response:%v =>", msg.ID, msg.Answers)

		dh.cLocker.RLock()
		conn, ok := dh.cache[msg.ID]
		if !ok {
			dh.cLocker.RUnlock()
			utils.LogInst().Warnf("======>>> no such[%d] cache item for response:%s", msg.ID, msg.GoString())
			continue
		}
		if len(msg.Answers) == 0 {
			dh.cLocker.RUnlock()
			utils.LogInst().Warnf("======>>> empty dns[%d] Answers", msg.ID)
			continue
		}

		dh.cLocker.RUnlock()
		conn.updateTime = time.Now()

		RInst().ParseDns(msg)

		_, err = conn.WriteFrom(buf[:n], addr)
		if err != nil {
			utils.LogInst().Errorf("======>>>dns proxy write back err:%s", err.Error())
			continue
		}
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
	dh.rLocker.RLock()
	peerUdp, ok := dh.redirectMap[target]
	if !ok {
		dh.rLocker.RUnlock()
		return
	}
	dh.rLocker.RUnlock()
	peerUdp.Close()
	dh.rLocker.Lock()
	delete(dh.redirectMap, target)
	dh.rLocker.Unlock()
}

func (dh *dnsHandler) forwardToTarget(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	id := udpID(conn.LocalAddr().String(), addr.String())
	dh.rLocker.RLock()
	peerUdp, ok := dh.redirectMap[id]
	if !ok {
		dh.rLocker.RUnlock()
		conn.Close()
		err := fmt.Errorf("no peer udp relay found for addr:%s", id)
		utils.LogInst().Warnf("======>>>udp relay app------>target err:=>%s", err.Error())
		return err
	}
	dh.rLocker.RUnlock()
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
	utils.LogInst().Debugf("======>>>dns[%d] questions:%v =>", msg.ID, msg.Questions)

	dh.cLocker.Lock()
	dh.cache[msg.ID] = &dnsConn{conn, time.Now()}
	dh.cLocker.Unlock()

	return nil
}
