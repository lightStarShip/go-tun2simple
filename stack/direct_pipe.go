package stack

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"io"
	"net"
	"time"
)

const (
	DialTimeOut   = 5 * time.Second
	SoftCloseTime = 2 * time.Second
)

func (s1 *stackV1) directRelay(conn net.Conn, target *net.TCPAddr) error {
	//targetConn, err := SafeConn("tcp", target.String(), s1.connSaver, DialTimeOut)
	targetConn, err := net.DialTCP("tcp", nil, target)
	if err != nil {
		_ = conn.Close()
		utils.LogInst().Errorf("======>>>tcp dial[%s] err:%v", target.String(), err)
		return err
	}
	utils.LogInst().Infof("======>>> direct relay for target:%s", target.String())
	go relay(conn, targetConn)
	go relay(targetConn, conn)
	return nil
}

func relay(src, dst net.Conn) {
	buf := utils.NewBytes(32 * 1024)
	defer utils.FreeBytes(buf)
	_, err := io.CopyBuffer(src, dst, buf)

	if err != nil {
		utils.LogInst().Warnf("======>>> direct relay finished:%s", err.Error())
	} else {
		utils.LogInst().Debugf("======>>> direct relay finished:[%s--->%s]===>[%s--->%s]",
			src.LocalAddr().String(),
			src.RemoteAddr().String(),
			dst.LocalAddr().String(),
			dst.RemoteAddr().String())
	}
	_ = src.Close()
	_ = dst.Close()
	//_ = src.SetDeadline(time.Now().Add(SoftCloseTime))
	//_ = dst.SetDeadline(time.Now().Add(SoftCloseTime))
}
