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
	c, err := net.Dial("tcp", target.String())
	if err != nil {
		utils.LogInst().Errorf("======>>>tcp dial[%s] err:%v", target.String(), err)
		return err
	}
	go h.handleInput(conn, c)
	go h.handleOutput(conn, c)
	utils.LogInst().Infof("new proxy connection for target: %s:%s", target.Network(), target.String())
	return nil
}
