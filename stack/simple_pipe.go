package stack

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"github.com/redeslab/go-simple/network"
	"github.com/redeslab/go-simple/node"
	"net"
)

func (s1 *stackV1) setupSimpleConn(nameTarget string) (net.Conn, error) {
	conn, err := SafeConn("tcp", s1.minerAddr, s1.connSaver, DialTimeOut)
	if err != nil {
		utils.LogInst().Errorf("======>>>SafeConn for[%s] server err :%v", nameTarget, err)
		return nil, err
	}
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	lvConn := network.NewLVConn(conn)

	iv := network.NewSalt()
	req := &node.SetupReq{
		IV:      *iv,
		SubAddr: s1.selfId,
		MTU:     s1.mtu,
	}
	jsonConn := &network.JsonConn{Conn: lvConn}
	buf := utils.NewBytes(utils.BufSize)
	defer utils.FreeBytes(buf)
	if err := jsonConn.SynBuffer(buf, req); err != nil {
		utils.LogInst().Errorf("======>>>SetupReq for[%s] server err :%v", nameTarget, err)
		return nil, err
	}

	aesConn, err := network.NewAesConn(lvConn, s1.aesKey, *iv)
	if err != nil {
		utils.LogInst().Errorf("======>>>NewAesConn for[%s] server err :%v", nameTarget, err)
		return nil, err
	}

	jsonConn = &network.JsonConn{Conn: aesConn}
	if err := jsonConn.SynBuffer(buf, &node.ProbeReq{
		Target: nameTarget,
	}); err != nil {
		utils.LogInst().Errorf("======>>>ProbeReq for[%s] server err :%v", nameTarget, err)
		return nil, err
	}

	return aesConn, nil
}
