package stack

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
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
	hip := net.ParseIP("45.253.43.45")
	boool := IPRuleInst().IsInnerIP(hip)
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
func TestLoadIP4(t *testing.T) {
	ptr, _ := net.LookupAddr("45.253.43.45")
	for _, ptrvalue := range ptr {
		fmt.Println(ptrvalue)
	}
}

func TestIPRange2(t *testing.T) {
	resp, err := http.Get("http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile("bypass.txt", body, 0644)
	if err != nil {
		t.Fatal(err)
	}
	/*
	   cat cn_raw.txt|grep "CN|ipv4" > cn_ipv4.txt

	*/
}

// go test -run TestIPRange1 --uid="BD"
func TestIPRange1(t *testing.T) {
	resp, err := http.Get("http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)

	file, err := os.Create("bypass3.txt")
	if err != nil {
		t.Fatal(err)
	}
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Println("--------------->finished----------->")
				break
			}
			t.Fatal(err)
		}
		strLine := string(line)
		subStrs := strings.Split(strLine, "|")
		if len(subStrs) != 7 {
			fmt.Println("xxxxxxxxxxx invalid: \t", strLine)
			continue
		}
		if subStrs[1] != uid || subStrs[2] != "ipv4" {
			fmt.Println("xxxxxxxxxxx not match\t", strLine)
			continue
		}

		ipNo, err := strconv.Atoi(subStrs[4])
		if err != nil {
			fmt.Println("xxxxxxxxxxx number convert failed\t", strLine)
			continue
		}

		ipPower := getPowerOfInt(ipNo)
		ret := fmt.Sprintf("%s/%d", subStrs[3], ipPower)
		_, err = file.WriteString(ret + "\n")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("+++++++right match:\t", ret)
	}
	file.Close()
}

func getPowerOfInt(size int) int {
	var mask = -1
	for size != 0 {
		size = size >> 1
		mask++
	}
	return 32 - mask
}

func TestIPRange3(t *testing.T) {
	fmt.Println(getPowerOfInt(64))
	fmt.Println(getPowerOfInt(1024))
	fmt.Println(getPowerOfInt(65536))
}

func TestNilChan(t *testing.T) {
	sig := make(chan struct{}, 1)

	//_, ok := <-sig
	//fmt.Println("ok=", ok)

	close(sig)
	_, ok := <-sig
	fmt.Println("ok=", ok)
}

/*
 cat cn_raw.txt|grep "CN|ipv4" > cn_ipv4.txt

*/
