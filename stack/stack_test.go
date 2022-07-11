package stack

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
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

func TestIpSubSet(t *testing.T) {
	ip, subNet, err := net.ParseCIDR("125.208.0.0/19")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("ip sub:", ip.String(), subNet.Mask.String())
	hip := net.ParseIP("125.209.222.59")
	maskIP := hip.Mask(subNet.Mask)

	fmt.Println("mip:", maskIP.String())
}

func TestLoadIP1(t *testing.T) {
	bts, err := ioutil.ReadFile("bypass2.txt")
	if err != nil {
		t.Fatal(err)
	}
	array := strings.Split(string(bts), "\n")
	for _, cidr := range array {
		ip, subNet, err := net.ParseCIDR(cidr)
		if err != nil {
			fmt.Println("=======>>> invalid  bypass cidr", cidr)
			continue
		}
		fmt.Println(ip.String(), subNet.Mask.String())
	}
}
func TestLoadIP2(t *testing.T) {
	bts, err := ioutil.ReadFile("bypass2.txt")
	if err != nil {
		t.Fatal(err)
	}
	IPRuleInst().LoadInners(string(bts))
	hip := net.ParseIP("125.209.222.59")
	boool := IPRuleInst().IsInnerIP(hip)
	fmt.Println("=======>>> IsInnerIP:->", hip, boool)
	hip = net.ParseIP("125.208.0.1")
	boool = IPRuleInst().IsInnerIP(hip)
	fmt.Println("=======>>> IsInnerIP:->", hip, boool)
	hip = net.ParseIP("125.208.31.255")
	boool = IPRuleInst().IsInnerIP(hip)
	fmt.Println("=======>>> IsInnerIP:->", hip, boool)
	hip = net.ParseIP("125.208.32.1")
	boool = IPRuleInst().IsInnerIP(hip)
	fmt.Println("=======>>> IsInnerIP:->", hip, boool)

}
func TestAesKey(t *testing.T) {
	_, err := hex.DecodeString(strings.ToLower("B9c0k2GRZDLn63i/REt0HAWCIR64zR6h48i87+XFz34="))
	if err != nil {
		t.Fatal(err)
	}
}
func TestLoadIP3(t *testing.T) {
	bts, err := ioutil.ReadFile("must_hit.txt")
	if err != nil {
		t.Fatal(err)
	}
	IPRuleInst().LoadInners(string(bts))
	hip := net.ParseIP("149.154.175.51")
	boool := IPRuleInst().IsInnerIP(hip)
	fmt.Println("=======>>> IsInnerIP:->", hip, boool)
}
