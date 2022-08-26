package stack

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
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
	dnsMap      map[uint16]*dnsConn
	redirectMap map[string]net.Conn
	expire      *time.Ticker
	ctx         context.Context
}

func newUdpHandler(saver ConnProtector, ctx context.Context) (core.UDPConnHandler, error) {

	handler := &dnsHandler{
		saver:       saver,
		ctx:         ctx,
		dnsMap:      make(map[uint16]*dnsConn),
		redirectMap: make(map[string]net.Conn),
		expire:      time.NewTicker(ExpireTime),
	}
	if err := handler.setupPivot(); err != nil {
		return nil, err
	}

	go handler.dnsWaitResponse()
	go handler.expireConn()
	return handler, nil
}

func udpID(src, dst string) string {
	return fmt.Sprintf("%s->%s", src, dst)
}

func (dh *dnsHandler) expireConn() {
	id := utils.GetGID()
	utils.LogInst().Infof("======>>> timer cleaner[%d] start success:", id)
	defer utils.LogInst().Infof("======>>> timer cleaner[%d] quit:", id)
	for {
		select {
		case <-dh.ctx.Done():
			utils.LogInst().Infof("======>>> timer cleaner[%d] quit by controller", id)
			return
		case tim, ok := <-dh.expire.C:
			if !ok {
				utils.LogInst().Warnf("======>>> timer cleaner[%d] exit", id)
				return
			}
			utils.LogInst().Infof("======>>> timer cleaner[%d] start[%s]:=>", id, tim.String())
			toDelete := make([]uint16, 0)
			for idx, conn := range dh.dnsMap {
				if tim.Sub(conn.updateTime) <= ExpireTime {
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
				delete(dh.dnsMap, idx)
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
		conn.Close()
		return err
	}
	id := udpID(conn.LocalAddr().String(), target.String())
	dh.rLocker.Lock()
	dh.redirectMap[id] = peerUdp
	dh.rLocker.Unlock()
	go dh.redirectWaitForRemote(conn, peerUdp, target)
	return nil
}

func (dh *dnsHandler) close() {
	utils.LogInst().Warnf("======>>>dns handler quit......")
	dh.cLocker.Lock()
	for _, conn := range dh.dnsMap {
		conn.Close()
	}
	dh.cLocker.Unlock()
	dh.pivot.Close()
}

func (dh *dnsHandler) setupPivot() error {
	if dh.pivot != nil {
		dh.pivot.Close()

	}
	bindAddr := &net.UDPAddr{IP: nil, Port: 0}
	pc, err := net.ListenUDP("udp4", bindAddr)
	if err != nil {
		utils.LogInst().Errorf("======>>>DNS ListenUDP err:=>%s", err.Error())
		return err
	}
	raw, err := pc.SyscallConn()
	if err != nil {
		utils.LogInst().Errorf("======>>>DNS SyscallConn err:=>%s", err.Error())
		return err
	}
	if err := raw.Control(dh.saver); err != nil {
		utils.LogInst().Errorf("======>>>DNS raw Control err:=>%s", err.Error())
		return err
	}
	dh.pivot = pc
	utils.LogInst().Debugf("======>>> create udp pivot at[%s] success:=>", pc.LocalAddr().String())
	return nil
}

func (dh *dnsHandler) dnsWaitResponse() {
	utils.LogInst().Infof("======>>> dns wait thread start work......")
	defer utils.LogInst().Infof("======>>> dns wait thread off work......")
	buf := utils.NewBytes(utils.BufSize)

	defer func() {
		dh.close()
		utils.FreeBytes(buf)
	}()

	for {
		select {
		case <-dh.ctx.Done():
			utils.LogInst().Infof("======>>> dns wait thread exit by app controller......")
			return
		default:
		}
		n, addr, err := dh.pivot.ReadFromUDP(buf)
		if err != nil {
			utils.LogInst().Errorf("======>>>udp pivot thread exit %v", err)
			return
		}
		msg := &dnsmessage.Message{}
		if err := msg.Unpack(buf[:n]); err != nil {
			utils.LogInst().Errorf("======>>>Unpack dns response err:%s\n%s\n", err.Error(), hex.EncodeToString(buf[:n]))
			continue
		}
		utils.LogInst().Debugf("======>>>dns[%d] response:%v =>", msg.ID, msg.Answers)

		dh.cLocker.RLock()
		conn, ok := dh.dnsMap[msg.ID]
		if !ok {
			dh.cLocker.RUnlock()
			utils.LogInst().Warnf("======>>> no such[%d] dnsMap item for response:%s", msg.ID, msg.GoString())
			continue
		}
		if len(msg.Answers) == 0 {
			dh.cLocker.RUnlock()
			utils.LogInst().Warnf("======>>> empty dns[%d] Answers from [%s]", msg.ID, addr.String())
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

func (dh *dnsHandler) redirectWaitForRemote(conn core.UDPConn, peerUdp net.Conn, target *net.UDPAddr) {
	buf := utils.NewBytes(utils.BufSize)
	defer utils.FreeBytes(buf)
	utils.LogInst().Warnf("======>>>prepare to read udp for target:=>%s", target.String())
	id := udpID(conn.LocalAddr().String(), target.String())
	defer dh.clearUdpRelay(id)
	defer conn.Close()
	for {
		select {
		case <-dh.ctx.Done():
			utils.LogInst().Infof("======>>> >udp relay thread exit by app controller......")
			return
		default:

		}
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

func (dh *dnsHandler) clearUdpRelay(id string) {
	dh.rLocker.RLock()
	peerUdp, ok := dh.redirectMap[id]
	if !ok {
		dh.rLocker.RUnlock()
		return
	}
	dh.rLocker.RUnlock()
	peerUdp.Close()
	dh.rLocker.Lock()
	delete(dh.redirectMap, id)
	dh.rLocker.Unlock()
}

func (dh *dnsHandler) redirectForwardToRemote(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
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
	dh.clearUdpRelay(id)
	conn.Close()
	utils.LogInst().Warnf("======>>>udp relay app------>target[id=%s] peer write err:=>%s", id, err.Error())
	return err
}

func (dh *dnsHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	utils.LogInst().Debugf("======>>>ReceiveTo %s------>>>%s", conn.LocalAddr().String(), addr)

	if addr.Port != COMMON_DNS_PORT {
		return dh.redirectForwardToRemote(conn, data, addr)
	}

	_, err := dh.pivot.WriteToUDP(data, addr)
	if err != nil {
		conn.Close()
		utils.LogInst().Errorf("======>>>dns forward err:%s\n%s\n", err.Error(), hex.EncodeToString(data))
		if err := dh.setupPivot(); err != nil {
			utils.LogInst().Errorf("======>>>restart dns pivot err:%s\n%s\n")
			return err
		}
		go dh.dnsWaitResponse()
		return err
	}

	msg := &dnsmessage.Message{}
	if err := msg.Unpack(data); err != nil {
		utils.LogInst().Errorf("======>>>Unpack dns request err:%s", err.Error(), hex.EncodeToString(data))
		return err
	}
	utils.LogInst().Infof("======>>>dns[%d] questions:%v =>", msg.ID, msg.Questions)

	dh.cLocker.Lock()
	dh.dnsMap[msg.ID] = &dnsConn{conn, time.Now()}
	dh.cLocker.Unlock()

	return nil
}
