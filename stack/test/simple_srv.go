package main

import (
	"encoding/json"
	"fmt"
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

	go io.Copy(conn, targetConn)
	fmt.Println("start working------>", sync.Target)
	io.Copy(targetConn, conn)
	targetConn.Close()
	conn.Close()
}
