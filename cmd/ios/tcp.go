package tun2Simple

import (
	"encoding/json"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/core"
	test "github.com/lightStarShip/go-tun2simple/stack"
	"github.com/lightStarShip/go-tun2simple/utils"
	"io"
	"net"
)

type tcpHandler struct {
}

type duplexConn interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

func newTCPHandler() core.TCPConnHandler {
	return &tcpHandler{}
}

func relay(src, dst net.Conn) {
	io.Copy(src, dst)
	src.Close()
	dst.Close()
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	var targetConn *net.TCPConn = nil

	if RInst().NeedProxy(target.IP.String()) {
		utils.LogInst().Infof("======>>>****** need a proxy for target:%s", target.String())
		c, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   net.ParseIP("149.248.37.162"),
			Port: 18888,
		})
		if err != nil {
			return err
		}

		if err := h.syncTarget(target.String(), c); err != nil {
			c.Close()
			utils.LogInst().Errorf("======>>>proxy sync target[%s] err:%v", target.String(), err)
			return err
		}

		targetConn = c
		utils.LogInst().Infof("======>>> proxy for target:%s", target.String())

	} else {
		c, err := net.DialTCP("tcp", nil, target)
		if err != nil {
			utils.LogInst().Errorf("======>>>tcp dial[%s] err:%v", target.String(), err)
			return err
		}
		targetConn = c
		utils.LogInst().Infof("======>>> direct relay for target:%s", target.String())
	}

	go relay(conn, targetConn)
	go relay(targetConn, conn)
	return nil
}

func (h *tcpHandler) syncTarget(target string, tConn *net.TCPConn) error {
	data, err := json.Marshal(&test.TestProxySync{
		Target: target,
	})
	if err != nil {
		return err
	}

	_, err = tConn.Write(data)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := tConn.Read(buf)
	if err != nil {
		return err
	}
	ack := &test.TestProxyAck{}
	err = json.Unmarshal(buf[:n], ack)
	if err != nil {
		return err

	}
	if ack.Msg != "OK" {
		return fmt.Errorf(ack.Msg)
	}
	return nil
}
