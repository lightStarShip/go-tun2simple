package stack

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"io"
	"net"
	"time"
)

const (
	DialTimeOut = 5 * time.Second
	MinMtuVal   = 1 << 15
)

func (s1 *stackV1) relay(src, dst net.Conn) {
	buf := utils.NewBytes(s1.mtu)
	defer utils.FreeBytes(buf)
	defer src.Close()
	defer dst.Close()

	_, err := io.CopyBuffer(src, dst, buf)
	if err != nil {
		utils.LogInst().Warnf("======>>> relay finalized by err:%s", err.Error())
		return
	}

	utils.LogInst().Debugf("======>>> relay finished:[%s--->%s]===>[%s--->%s]",
		src.LocalAddr().String(),
		src.RemoteAddr().String(),
		dst.LocalAddr().String(),
		dst.RemoteAddr().String())
}
