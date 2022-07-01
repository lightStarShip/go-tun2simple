package tun2Simple

import (
	"errors"
	"github.com/lightStarShip/go-tun2simple/core"
	"github.com/lightStarShip/go-tun2simple/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"sync"
)

type dnsHandler struct {
	sync.Mutex
	pivot *net.UDPConn
	cache map[uint16]core.UDPConn
}

func NewDnsHandler() core.UDPConnHandler {
	bindAddr := &net.UDPAddr{IP: nil, Port: 0}
	pc, err := net.ListenUDP("udp4", bindAddr)
	if err != nil {
		panic(err)
	}
	handler := &dnsHandler{
		pivot: pc,
		cache: make(map[uint16]core.UDPConn),
	}
	go handler.waitResponse()
	return handler
}

func (dh *dnsHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	utils.LogInst().Debugf("======>>>Connect:%s------>>>%s", conn.LocalAddr().String(), target.String())
	if target.Port != COMMON_DNS_PORT {
		utils.LogInst().Errorf("======>>>Cannot handle non-DNS packet port:%s", target.String())
		return errors.New("can not handle non-DNS packet")
	}
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
			utils.LogInst().Errorf("======>>>Unpack dns response err:%s", err.Error())
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

func (dh *dnsHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	utils.LogInst().Debugf("======>>>ReceiveTo %s------>>>%s", conn.LocalAddr().String(), addr)
	msg := &dnsmessage.Message{}
	if err := msg.Unpack(data); err != nil {
		utils.LogInst().Errorf("======>>>Unpack dns request err:%s", err.Error())
		return err
	}
	utils.LogInst().Debugf("======>>>dns[%d] questions:%s =>", msg.ID, msg.Questions)

	dh.Lock()
	dh.cache[msg.ID] = conn
	dh.Unlock()

	_, err := dh.pivot.WriteToUDP(data, addr)
	if err != nil {
		utils.LogInst().Errorf("======>>>dns forward err:%s", err.Error())
		return err
	}

	return nil
}
