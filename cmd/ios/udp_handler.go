package tun2Simple

import (
	"errors"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

type dnsHandler struct {
}

func NewDnsHandler() core.UDPConnHandler {
	return &dnsHandler{}
}

func (dh *dnsHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	console("======>>>Connect:", conn.LocalAddr().String(), target.String())
	if target.Port != COMMON_DNS_PORT {
		console("======>>>Cannot handle non-DNS packet")
		return errors.New("Cannot handle non-DNS packet")
	}
	return nil
}

func (dh *dnsHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	console("======>>>ReceiveTo:", conn.LocalAddr().String(), addr)

	if len(data) < dnsHeaderLength {
		console("======>>>Received malformed DNS query")
		return fmt.Errorf("received malformed DNS query")
	}
	msg := &dnsmessage.Message{}
	if err := msg.Unpack(data); err != nil {
		console("======>>>Unpack dns err:", err.Error())
		return err
	}

	for idx, question := range msg.Questions {
		console("======>>>dns query:=>", idx, question.GoString())
	}
	//_, err := conn.WriteFrom(data, addr)
	//if err != nil {
	//	console("======>>>conn write from err:", err.Error())
	//}
	//return err
	return nil
}
