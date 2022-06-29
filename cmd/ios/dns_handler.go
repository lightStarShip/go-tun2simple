package tun2Simple

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

type dnsHandler struct {
}

func NewDnsHandler() core.UDPConnHandler {
	handler := &dnsHandler{}
	return handler
}

func (dh *dnsHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	console("======>>>Connect:", conn.LocalAddr().String(), target.String())
	if target.Port != COMMON_DNS_PORT {
		console("======>>>Cannot handle non-DNS packet")
		return errors.New("can not handle non-DNS packet")
	}
	return nil
}

func (dh *dnsHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	console("======>>>ReceiveTo:", conn.LocalAddr().String(), addr)

	if len(data) < dnsHeaderLength {
		console("======>>>Received malformed DNS query")
		return fmt.Errorf("received malformed DNS query")
	}

	netUdpConn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		console("======>>>forward dns err:", err.Error())
		return err
	}
	defer netUdpConn.Close()

	_, err = netUdpConn.Write(data)
	if err != nil {
		console("======>>>dns forward err:", err.Error())
		return err
	}

	resBuff := make([]byte, 4096) //TODO size ===lws
	n, err := bufio.NewReader(netUdpConn).Read(resBuff)
	if err != nil {
		console("======>>>dns read dns respnose err:", err.Error())
		return err
	}
	_, err = conn.WriteFrom(resBuff[:n], addr)
	if err != nil {
		console("======>>>dns proxy write back err:", err.Error())
		return err
	}
	msg := &dnsmessage.Message{}
	if err := msg.Unpack(resBuff[:n]); err != nil {
		console("======>>>Unpack dns err:", err.Error())
		return err
	}
	console("======>>>dns query:=>", n, msg.ID, msg.GoString())
	return nil
}
