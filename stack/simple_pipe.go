package stack

import (
	"fmt"
	"github.com/lightStarShip/go-tun2simple/utils"
	"github.com/redeslab/go-simple/network"
	"github.com/redeslab/go-simple/node"
	"net"
)

func (s1 *stackV1) SimpleRelay(conn net.Conn, target *net.TCPAddr, matched string) error {
	utils.LogInst().Infof("======>>>****** prepare to proxy for target:%s", target.String())

	nameTarget := fmt.Sprintf("%s:%d", matched, target.Port)
	tarConn, err := s1.setupSimpleConn(nameTarget)
	if err != nil {
		_ = conn.Close()
		utils.LogInst().Errorf("======>>>proxy sync target[%s=>%s] err:%v", target.String(), nameTarget, err)
		return err
	}
	utils.LogInst().Infof("======>>> proxy for target:[%s=>%s]", target.String(), nameTarget)

	go relay(conn, tarConn)
	go relay(tarConn, conn)
	return nil
}

func (s1 *stackV1) setupSimpleConn(nameTarget string) (net.Conn, error) {
	conn, err := SafeConn("tcp", s1.minerAddr, s1.connSaver, DialTimeOut)
	if err != nil {
		utils.LogInst().Errorf("======>>>proxy for[%s] server err :%v", nameTarget, err)
		return nil, err
	}
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	lvConn := network.NewLVConn(conn)

	iv := network.NewSalt()
	req := &node.SetupReq{
		IV:      *iv,
		SubAddr: s1.selfId,
	}
	jsonConn := &network.JsonConn{Conn: lvConn}
	buf := utils.NewBytes(1024)
	defer utils.FreeBytes(buf)
	if err := jsonConn.SynBuffer(buf, req); err != nil {
		return nil, err
	}
	aesConn, err := network.NewAesConn(lvConn, s1.aesKey, *iv)
	if err != nil {
		return nil, err
	}
	jsonConn = &network.JsonConn{Conn: aesConn}
	if err := jsonConn.SynBuffer(buf, &node.ProbeReq{
		Target: nameTarget,
	}); err != nil {
		return nil, err
	}

	return aesConn, nil
}
