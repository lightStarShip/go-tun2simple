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
	buf := utils.NewBytes(32 * 1024)
	defer utils.FreeBytes(buf)
	io.CopyBuffer(src, dst, buf)
	src.Close()
	dst.Close()
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	var targetConn *net.TCPConn = nil

	matched := RInst().NeedProxy(target.IP.String())
	if len(matched) > 0 {
		utils.LogInst().Infof("======>>>****** prepare to proxy for target:%s", target.String())
		c, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   net.ParseIP("149.248.37.162"),
			Port: 18888,
		})
		if err != nil {
			utils.LogInst().Errorf("======>>>proxy for[%s] server err :%v", target.String(), err)
			return err
		}
		nameTarget := fmt.Sprintf("%s:%d", matched, target.Port)
		if err := h.syncTarget(nameTarget, c); err != nil {
			c.Close()
			conn.Close()
			utils.LogInst().Errorf("======>>>proxy sync target[%s=>%s] err:%v", target.String(), nameTarget, err)
			return err
		}

		targetConn = c
		utils.LogInst().Infof("======>>> proxy for target:[%s=>%s]", target.String(), nameTarget)

	} else {
		c, err := net.DialTCP("tcp", nil, target)
		if err != nil {
			conn.Close()
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
	buf := utils.NewBytes(utils.BufSize)
	defer utils.FreeBytes(buf)
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
