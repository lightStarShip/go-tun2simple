package stack

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"github.com/redeslab/go-simple/network"
	"io"
	"net"
	"time"
)

const (
	DialTimeOut = 5 * time.Second
)

func (s1 *stackV1) relay(conn, target net.Conn) {
	go relay(conn, target)
	relay(target, conn)

}

func relay(src, dst net.Conn) {
	buf := utils.NewBytes(network.MTU)
	defer utils.FreeBytes(buf)
	defer src.Close()
	defer dst.Close()

	_, err := io.CopyBuffer(src, dst, buf)
	if err != nil {
		utils.LogInst().Warnf("======>>> direct relay finalized:%s", err.Error())
		return
	}

	utils.LogInst().Debugf("======>>> direct relay finished:[%s--->%s]===>[%s--->%s]",
		src.LocalAddr().String(),
		src.RemoteAddr().String(),
		dst.LocalAddr().String(),
		dst.RemoteAddr().String())
}
