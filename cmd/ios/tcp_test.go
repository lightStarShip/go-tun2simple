package tun2Simple

import (
	"encoding/json"
	"fmt"
	test "github.com/lightStarShip/go-tun2simple/stack"
	"net"
	"testing"
)

func TestTcpDial(t *testing.T) {
	c, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP("149.248.37.162"),
		Port: 18888,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(&test.TestProxySync{
		Target: "www.baidu.com:443",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	ack := &test.TestProxyAck{}
	err = json.Unmarshal(buf[:n], ack)
	if err != nil {
		t.Fatal(err)

	}
	if ack.Msg != "OK" {
		t.Fatal("---------> not ok", ack.Msg)
	}

	c.Write(data)
	c.Write(data)
	c.Write(data)
	c.Write(data)
}

func TestDnsLookup(t *testing.T) {
	fmt.Println(net.LookupHost("rr4---sn-nx57ynld.googlevideo.com."))
}
