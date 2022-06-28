package tun2Simple

import (
	"errors"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"golang.org/x/net/dns/dnsmessage"
	"io"
	"net"
	"runtime/debug"
	"time"
)

func init() {
	// Apple VPN extensions have a memory limit of 15MB. Conserve memory by increasing garbage
	// collection frequency and returning memory to the OS every minute.
	debug.SetGCPercent(10)
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for range ticker.C {
			fmt.Println("======>>> release memory for ios")
			debug.FreeOSMemory()
		}
	}()
}

const (
	COMMON_DNS_PORT = 53
	dnsHeaderLength = 12
	dnsMaskQr       = uint8(0x80)
	dnsMaskTc       = uint8(0x02)
	dnsMaskRcode    = uint8(0x0F)
)

var _iosApp *outlinetunnel = nil

type Tunnel interface {
	Write(data []byte) (int, error)
}

type outlinetunnel struct {
	lwipStack core.LWIPStack
	dev       TunnelDev
}
type TunnelDev interface {
	io.WriteCloser
	Log(s string)
}

func console(a ...any) {
	log := fmt.Sprintln(a)
	_iosApp.dev.Log(log)
}

func NewTunnel(tunWriter TunnelDev) (Tunnel, error) {
	if tunWriter == nil {
		return nil, errors.New("Must provide a TUN writer")
	}

	core.RegisterOutputFn(func(data []byte) (int, error) {
		console("--------------", string(data))
		return tunWriter.Write(data)
	})
	lwipStack := core.NewLWIPStack()
	t := &outlinetunnel{
		lwipStack,
		tunWriter}
	core.RegisterTCPConnHandler(t)
	core.RegisterUDPConnHandler(t)
	_iosApp = t
	return t, nil
}

func (t *outlinetunnel) Write(data []byte) (int, error) {
	return t.lwipStack.Write(data)
}

func (t *outlinetunnel) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	console("======>>>Connect:", conn.LocalAddr().String(), target.String())
	if target.Port != COMMON_DNS_PORT {
		console("======>>>Cannot handle non-DNS packet")
		return errors.New("Cannot handle non-DNS packet")
	}
	return nil
}

func (t *outlinetunnel) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
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
		console("======>>>dns query:=>", idx, question.Name)
	}

	//
	////  DNS Header
	////  0  1  2  3  4  5  6  7  0  1  2  3  4  5  6  7
	////  +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	////  |                      ID                       |
	////  +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	////  |QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
	////  +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	////  |                    QDCOUNT                    |
	////  +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	////  |                    ANCOUNT                    |
	////  +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	////  |                    NSCOUNT                    |
	////  +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	////  |                    ARCOUNT                    |
	////  +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//// Set response and truncated bits
	//data[2] |= dnsMaskQr | dnsMaskTc
	//// Set response code to 'no error'.
	//data[3] &= ^dnsMaskRcode
	//// Set ANCOUNT to QDCOUNT. This is technically incorrect, since the response does not
	//// include an answer. However, without it some DNS clients (i.e. Windows 7) do not retry
	//// over TCP.
	//var qdcount = binary.BigEndian.Uint16(data[4:6])
	//binary.BigEndian.PutUint16(data[6:], qdcount)
	_, err := conn.WriteFrom(data, addr)
	if err != nil {
		console("======>>>conn write from err:", err.Error())
	}
	return err
}

func (t *outlinetunnel) Handle(conn net.Conn, target *net.TCPAddr) error {
	console("======>>>Handle implement me:", conn.LocalAddr(), target.String())
	return nil
}
