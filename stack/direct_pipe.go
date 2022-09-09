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

func (s1 *stackV1) relayForProxy(src, dst net.Conn) {
	buf := utils.NewBytes(s1.mtu)
	defer utils.FreeBytes(buf)

	for {
		n, err := src.Read(buf)
		if err != nil {
			utils.LogInst().Warnf("======>>> read from source err:%s", err.Error())
			break
		}
		_, err = dst.Write(buf[:n])
		if err != nil {
			utils.LogInst().Warnf("======>>> write to dst err:%s", err.Error())
			break
		}
	}

	defer src.Close()
	defer dst.Close()

	utils.LogInst().Debugf("======>>>proxy relay finished:[%s--->%s]===>[%s--->%s]",
		src.LocalAddr().String(),
		src.RemoteAddr().String(),
		dst.LocalAddr().String(),
		dst.RemoteAddr().String())
}

func (s1 *stackV1) relay(src, dst net.Conn) {
	buf := utils.NewBytes(s1.mtu)
	defer utils.FreeBytes(buf)
	defer src.Close()
	defer dst.Close()

	_, err := io.CopyBuffer(src, dst, buf)
	if err != nil {
		utils.LogInst().Warnf("======>>> direct relay finalized by err:%s", err.Error())
		return
	}

	utils.LogInst().Debugf("======>>>  direct relay finished:[%s--->%s]===>[%s--->%s]",
		src.LocalAddr().String(),
		src.RemoteAddr().String(),
		dst.LocalAddr().String(),
		dst.RemoteAddr().String())
}

func (s1 *stackV1) upStream(isProxy bool, appConn, proxyConn net.Conn) {
	buf := utils.NewBytes(s1.mtu)
	defer utils.FreeBytes(buf)
	for {
		no, err := appConn.Read(buf)
		if no == 0 {
			if err != io.EOF {
				utils.LogInst().Warnf("======>>>[proxy=%t]read:app---->proxy err=>%s left:%d local is :%s", isProxy, err, no, appConn.LocalAddr().String())
			} else {
				utils.LogInst().Debugf("======>>>[proxy=%t]read:app---->proxy EOF local is :%s", isProxy, appConn.LocalAddr().String())
			}
			_ = appConn.SetDeadline(time.Now().Add(time.Second * 5))
			return
		}
		_, err = proxyConn.Write(buf[:no])
		if err != nil {
			proxyConn.Close()
			utils.LogInst().Warnf("======>>>[proxy=%t]write: app---->proxy err=>%s remote is:%s", isProxy, err, proxyConn.RemoteAddr().String())
			return
		}
		utils.LogInst().Debugf("======>>>[proxy=%t]upStream: app---->proxy data:%d ", isProxy, no)
	}
}

func (s1 *stackV1) downStream(isProxy bool, appConn, proxyConn net.Conn) {
	buf := utils.NewBytes(s1.mtu)
	defer utils.FreeBytes(buf)
	for {
		no, err := proxyConn.Read(buf)
		if no == 0 {
			if err != io.EOF {
				utils.LogInst().Warnf("======>>>[proxy=%t]read: app<----proxy err=>%s  remote is:%s", isProxy, err, proxyConn.RemoteAddr().String())
			} else {
				utils.LogInst().Debugf("======>>>[proxy=%t]read: app<----proxy EOF  remote is:%s", isProxy, proxyConn.RemoteAddr().String())
			}
			_ = proxyConn.SetDeadline(time.Now().Add(time.Second * 5))
			return
		}

		writeNo, err := appConn.Write(buf[:no])
		if err != nil {
			appConn.Close()
			utils.LogInst().Warnf("======>>>[proxy=%t]write app<----proxy err:%s left=%d local is :%s", isProxy, err, no, appConn.LocalAddr().String())
			break
		}

		utils.LogInst().Debugf("======>>>[proxy=%t]read: app<----proxy data:%d written:%d", isProxy, no, writeNo)
	}
}
