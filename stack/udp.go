package stack

import (
	"context"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"net"
	"sync"
)

type udpPivot struct {
	rLocker     sync.RWMutex
	saver       ConnProtector
	pivot       *net.UDPConn
	redirectMap map[string]net.Conn
	ctx         context.Context
}

func newUdpRelay(saver ConnProtector) core.UDPConnHandler {
	utils.LogInst().Infof("======>>>[UDP] new udp pivot :")
	up := &udpPivot{
		saver:       saver,
		redirectMap: make(map[string]net.Conn),
	}
	return up
}

func (up *udpPivot) redirectWaitForRemote(conn core.UDPConn, peerUdp net.Conn, target *net.UDPAddr) {
	buf := utils.NewBytes(utils.BufSize)
	defer utils.FreeBytes(buf)
	utils.LogInst().Warnf("======>>>[UDP]reading from target:=>%s", target.String())
	id := udpID(conn.LocalAddr().String(), target.String())
	defer up.clearUdpRelay(id)
	defer conn.Close()
	for {
		n, err := peerUdp.Read(buf)
		if err != nil {
			utils.LogInst().Warnf("======>>>[UDP] relay app<------target err:=>%s", err.Error())
			return
		}
		_, err = conn.WriteFrom(buf[:n], target)
		if err != nil {
			utils.LogInst().Warnf("======>>>[UDP] relay app<------target err:=>%s", err.Error())
			return
		}
	}
}

func (up *udpPivot) clearUdpRelay(id string) {
	up.rLocker.RLock()
	peerUdp, ok := up.redirectMap[id]
	if !ok {
		up.rLocker.RUnlock()
		return
	}
	up.rLocker.RUnlock()
	peerUdp.Close()
	up.rLocker.Lock()
	delete(up.redirectMap, id)
	up.rLocker.Unlock()
}
func (up *udpPivot) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	utils.LogInst().Debugf("======>>>udp relay Connect:%s------>>>%s", conn.LocalAddr().String(), target.String())

	peerUdp, err := SafeConn("udp", target.String(), up.saver, DialTimeOut)
	if err != nil {
		conn.Close()
		return err
	}
	id := udpID(conn.LocalAddr().String(), target.String())
	up.rLocker.Lock()
	up.redirectMap[id] = peerUdp
	up.rLocker.Unlock()
	go up.redirectWaitForRemote(conn, peerUdp, target)
	return nil
}

func (up *udpPivot) close() {
	utils.LogInst().Warnf("======>>>dns handler quit......")

	up.rLocker.Lock()
	up.redirectMap = make(map[string]net.Conn)
	up.rLocker.Unlock()

	up.pivot.Close()
}

func (up *udpPivot) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	utils.LogInst().Debugf("======>>>[UDP]ReceiveTo %s------>>>%s", conn.LocalAddr().String(), addr)

	id := udpID(conn.LocalAddr().String(), addr.String())
	up.rLocker.RLock()
	peerUdp, ok := up.redirectMap[id]
	if !ok {
		up.rLocker.RUnlock()
		conn.Close()
		err := fmt.Errorf("no peer udp relay found for addr:%s", id)
		utils.LogInst().Warnf("======>>>[UDP]relay app------>target err:=>%s", err.Error())
		return err
	}
	up.rLocker.RUnlock()
	_, err := peerUdp.Write(data)
	if err == nil {
		return nil
	}
	up.clearUdpRelay(id)
	conn.Close()
	utils.LogInst().Warnf("======>>>[UDP] relay app------>target[id=%s] peer write err:=>%s", id, err.Error())
	return err
}
