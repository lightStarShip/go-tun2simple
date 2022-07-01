package tun2Simple

import (
	"github.com/lightStarShip/go-tun2simple/core"
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

func (h *tcpHandler) handleInput(conn net.Conn, input io.ReadCloser) {
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

func (h *tcpHandler) handleOutput(conn net.Conn, output io.WriteCloser) {
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

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	var targetConn *net.TCPConn = nil
	if RInst().NeedProxy(target.IP.String()) {
		utils.LogInst().Infof("======>>>****** need a proxy for target:%s", target.String())
		c, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   net.ParseIP(""),
			Port: 18888,
		})
		if err != nil {
			return err
		}
		if err := h.syncTarget(c); err != nil {
			return err
		}
		targetConn = c

	} else {
		utils.LogInst().Infof("======>>> direct relay for target:%s", target.String())
		c, err := net.DialTCP("tcp", nil, target)
		if err != nil {
			utils.LogInst().Errorf("======>>>tcp dial[%s] err:%v", target.String(), err)
			return err
		}
		targetConn = c
	}

	go h.handleInput(conn, targetConn)
	go h.handleOutput(conn, targetConn)
	return nil
}

func (h *tcpHandler) syncTarget(tConn *net.TCPConn) error {
	return nil
}
