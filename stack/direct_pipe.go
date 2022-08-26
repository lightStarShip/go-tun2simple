package stack

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"io"
	"net"
	"time"
)

const (
	DialTimeOut = 8 * time.Second
	MinMtuVal   = 1 << 12
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

func (s1 *stackV1) upStream(isProxy bool, appConn, proxyConn net.Conn) {
	buf := utils.NewBytes(s1.mtu)
	defer utils.FreeBytes(buf)
	defer appConn.Close()
	for {
		no, err := appConn.Read(buf)
		if no == 0 {
			if err != io.EOF {
				utils.LogInst().Warnf("======>>>[proxy=%t]read:app---->proxy err=>%s left:%d", isProxy, err, no)
			} else {
				utils.LogInst().Debugf("======>>>[proxy=%t]read:app---->proxy EOF", isProxy)
			}
			return
		}
		_, err = proxyConn.Write(buf[:no])
		if err != nil {
			utils.LogInst().Warnf("======>>>[proxy=%t]write: app---->proxy err=>%s", isProxy, err)
			return
		}
		utils.LogInst().Debugf("======>>>[proxy=%t]upStream: app---->proxy data:%d ", isProxy, no)
	}
}

func (s1 *stackV1) downStream(isProxy bool, appConn, proxyConn net.Conn) {
	buf := utils.NewBytes(s1.mtu)
	defer utils.FreeBytes(buf)
	defer proxyConn.Close()
	for {
		no, err := proxyConn.Read(buf)
		if no == 0 {
			if err != io.EOF {
				utils.LogInst().Warnf("======>>>[proxy=%t]read: app<----proxy err=>%s", isProxy, err)
			} else {
				utils.LogInst().Debugf("======>>>[proxy=%t]read: app<----proxy EOF ", isProxy)
			}
			_ = appConn.SetDeadline(time.Now().Add(time.Second * 5))
			return
		}

		writeNo, err := appConn.Write(buf[:no])
		if err != nil {
			utils.LogInst().Warnf("======>>>[proxy=%t]write app<----proxy err:%s left=%d", isProxy, err, no)
			break
		}

		utils.LogInst().Debugf("======>>>[proxy=%t]read: app<----proxy data:%d written:%d", isProxy, no, writeNo)
	}
}
