package stack

import (
	"encoding/json"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	"io"
	"net"
)

func SimpleStack() {
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
	Msg string
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
			Msg: err.Error(),
		})
		conn.Write(data)
		fmt.Println(err)
		return
	}

	data, _ := json.Marshal(&TestProxyAck{
		Msg: "OK",
	})
	conn.Write(data)

	go handleInput(conn, targetConn)
	fmt.Println("start working------>", sync.Target)
	handleOutput(conn, targetConn)
}
