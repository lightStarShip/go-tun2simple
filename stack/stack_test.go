package stack

import (
	"fmt"
	"net"
	"testing"
)

func TestTcpDial(t *testing.T) {

}

func TestDnsLookup(t *testing.T) {
	fmt.Println(net.LookupHost(uid))
}

func TestDnsLookIP(t *testing.T) {
	fmt.Println(net.LookupAddr(uid))
}
