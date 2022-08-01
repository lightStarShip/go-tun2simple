package stack

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/lightStarShip/go-tun2simple/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"testing"
)

var p1 = "450000500019000040116e7b010101010a0000080035cbd8003cffdc6bba01000001000000000000136f617574686163636f756e746d616e616765720a676f6f676c656170697303636f6d0000010001"
var p2 = "4500004b0033000040116058080808080a0000080035f17d00370f536367010000010000000000000f636f6d6d6e61742d6d61696e2d676303657373056170706c6503636f6d0000010001"

func init() {
	utils.LogInst().InitParam(utils.DEBUG, func(msg string, args ...any) {
		fmt.Printf(msg, args...)
	})
}

func TestUnpackDns(t *testing.T) {

	buff, err := hex.DecodeString(p1)
	if err != nil {
		t.Fatal(err)
	}

	packet := gopacket.NewPacket(buff, layers.LayerTypeIPv4, gopacket.Default)

	var udp *layers.UDP = nil
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp = udpLayer.(*layers.UDP)
	} else {
		fmt.Println("======>>>Unpack dns err:")
		return
	}

	payload := udp.Payload

	fmt.Println(hex.EncodeToString(payload))

	msg := &dnsmessage.Message{}
	if err := msg.Unpack(payload); err != nil {
		t.Fatal(err)
	}
	fmt.Println(msg.GoString())

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.ParseIP("8.8.8.8"),
		Port: 53,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write(payload)

	if err != nil {
		t.Fatal(err)
	}
	p := make([]byte, 2048)
	n, err := bufio.NewReader(conn).Read(p)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println()
	msg2 := &dnsmessage.Message{}
	if err := msg2.Unpack(p[:n]); err != nil {
		t.Fatal(err)
	}
	fmt.Println(msg2.GoString())
}

func TestUnpackUdp(t *testing.T) {

	buff, err := hex.DecodeString("ce0401000001000000000000117272352d2d2d736e2d6f3039377a6e73640b676f6f676c65766964656f03636f6d0000010001")
	if err != nil {
		t.Fatal(err)
	}

	msg := &dnsmessage.Message{}
	if err := msg.Unpack(buff); err != nil {
		t.Fatal(err)
	}
	fmt.Println(msg.GoString())
}

func TestGoogleDns(t *testing.T) {

	buff, err := hex.DecodeString("ce0401000001000000000000117272352d2d2d736e2d6f3039377a6e73640b676f6f676c65766964656f03636f6d0000010001")
	if err != nil {
		t.Fatal(err)
	}

	msg := &dnsmessage.Message{}
	if err := msg.Unpack(buff); err != nil {
		t.Fatal(err)
	}
	fmt.Println(msg.GoString())

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.ParseIP("8.8.8.8"),
		Port: 53,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write(buff)

	if err != nil {
		t.Fatal(err)
	}
	p := make([]byte, 2048)
	n, err := bufio.NewReader(conn).Read(p)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println()
	msg2 := &dnsmessage.Message{}
	if err := msg2.Unpack(p[:n]); err != nil {
		t.Fatal(err)
	}
	fmt.Println(msg2.GoString())
}
