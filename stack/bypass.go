package stack

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"net"
	"strings"
	"sync"
)

type ByPassIPs struct {
	ipMask map[string]net.IPMask
	sync.RWMutex
	global bool
}

var _instance *ByPassIPs
var once sync.Once

func ByPassInst() *ByPassIPs {
	once.Do(func() {
		_instance = &ByPassIPs{
			ipMask: make(map[string]net.IPMask),
		}
	})
	return _instance
}

func (bp *ByPassIPs) Load(IPS string) {
	bp.ipMask = make(map[string]net.IPMask)

	array := strings.Split(IPS, "\n")
	for _, cidr := range array {
		ip, subNet, err := net.ParseCIDR(cidr)
		if err != nil {
			utils.LogInst().Debugf("=======>>> invalid  bypass cidr %s\n", cidr)
			continue
		}
		bp.ipMask[ip.String()] = subNet.Mask
	}
	utils.LogInst().Infof("=======>>> Total bypass :%d \n", len(bp.ipMask))
}

func (bp *ByPassIPs) IsInnerIP(srcIP net.IP) bool {

	bp.RLock()
	defer bp.RUnlock()

	if len(bp.ipMask) == 0 {
		utils.LogInst().Debugf("=======>>> no ip rule is used\n")
		return true
	}

	if bp.global {
		return false
	}

	for mip, mask := range bp.ipMask {
		maskIP := srcIP.Mask(mask).String()
		if maskIP == mip {
			utils.LogInst().Debugf("=======>>> IsInnerIP success ip:%s->mip:%s mask:%s\n", srcIP, mip, mask.String())
			return true
		}
	}

	return false
}

func (bp *ByPassIPs) SetGlobal(g bool) {
	bp.Lock()
	defer bp.Unlock()
	bp.global = g
}

func (bp *ByPassIPs) IsGlobal() bool {
	bp.RLock()
	defer bp.RUnlock()
	return bp.global
}
