package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"io"
	"net"
)

var (
	tcpRaw = []string{
		"450000410445000040115c500a00000808080808f1a10035002d46d662a9010000010000000000000f6170702d6d6561737572656d656e7403636f6d0000010001",
		"4500003c18810000401148190a00000808080808d54500350028ee46931d010000010000000000000377777706676f6f676c6503636f6d0000010001",
		"450000407fee00004011e0a70a00000808080808c9410035002cf257b2fd01000001000000000000037777770a676f6f676c656170697303636f6d0000410001",
		"45000040647600004011fc1f0a00000808080808d2290035002cb401e8ab01000001000000000000037777770a676f6f676c656170697303636f6d0000010001",
		"45000050987000004011c8150a00000808080808d4610035003c4372efc5010000010000000000000c696e732d7232337473757566036961730d74656e63656e742d636c6f7564036e65740000010001",
		"45000040af9b00004011b0fa0a00000808080808d4ab0035002cf11eb9ec010000010000000000000872656465736c61620667697468756202696f0000410001",
	}
)

func decode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func main() {
	src, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   nil,
		Port: 18888,
	})
	if err != nil {
		panic(err)
	}

	for {
		conn, err := src.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go relay(conn)
	}
}

type TestProxySync struct {
	Target string
}

type TestProxyAck struct {
	msg string
}
type duplexConn interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

func handleInput(conn net.Conn, input io.ReadCloser) {
	defer func() {
		if tcpConn, ok := conn.(core.TCPConn); ok {
			tcpConn.CloseWrite()
		} else {
			conn.Close()
		}
		if tcpInput, ok := input.(duplexConn); ok {
			tcpInput.CloseRead()
		} else {
			input.Close()
		}
	}()

	io.Copy(conn, input)
}

func handleOutput(conn net.Conn, output io.WriteCloser) {
	defer func() {
		if tcpConn, ok := conn.(core.TCPConn); ok {
			tcpConn.CloseRead()
		} else {
			conn.Close()
		}
		if tcpOutput, ok := output.(duplexConn); ok {
			tcpOutput.CloseWrite()
		} else {
			output.Close()
		}
	}()

	io.Copy(output, conn)
}

func relay(conn *net.TCPConn) {
	buf := make([]byte, 1<<20)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	sync := &TestProxySync{}
	if err := json.Unmarshal(buf[:n], sync); err != nil {
		panic(err)
	}
	fmt.Println("new conn------>", sync.Target)

	targetConn, err := net.Dial("tcp", sync.Target)
	if err != nil {
		data, _ := json.Marshal(&TestProxyAck{
			msg: err.Error(),
		})
		conn.Write(data)
		fmt.Println(err)
		return
	}

	data, _ := json.Marshal(&TestProxyAck{
		msg: "OK",
	})
	conn.Write(data)

	go handleInput(conn, targetConn)
	handleOutput(conn, targetConn)
}
