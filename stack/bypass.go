package stack

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"net"
	"strings"
	"sync"
)

type IpRules struct {
	innerIpMasks map[string]net.IPMask
	mustHitMasks map[string]net.IPMask
	sync.RWMutex
	global bool
}

var _instance *IpRules
var once sync.Once

func IPRuleInst() *IpRules {
	once.Do(func() {
		_instance = &IpRules{
			innerIpMasks: make(map[string]net.IPMask),
		}
	})
	return _instance
}

func (bp *IpRules) LoadInners(innerIPs string) {
	bp.innerIpMasks = make(map[string]net.IPMask)

	array := strings.Split(innerIPs, "\n")
	for _, cidr := range array {
		ip, subNet, err := net.ParseCIDR(cidr)
		if err != nil {
			utils.LogInst().Debugf("=======>>> invalid  bypass cidr [%s]\n", cidr)
			continue
		}
		bp.innerIpMasks[ip.String()] = subNet.Mask
	}
	utils.LogInst().Infof("=======>>> Total bypass :%d \n", len(bp.innerIpMasks))
}

func (bp *IpRules) IsInnerIP(srcIP net.IP) bool {

	bp.RLock()
	defer bp.RUnlock()

	if len(bp.innerIpMasks) == 0 {
		utils.LogInst().Debugf("=======>>> empty bypass ip rule is used [%s]\n", srcIP.String())
		return true
	}

	if bp.global {
		return false
	}

	for mip, mask := range bp.innerIpMasks {
		maskIP := srcIP.Mask(mask).String()
		if maskIP == mip {
			utils.LogInst().Debugf("=======>>> IsInnerIP success ip:%s->mip:%s mask:%s\n", srcIP, mip, mask.String())
			return true
		}
	}

	return false
}

func (bp *IpRules) IsMustHits(srcIP net.IP) bool {

	bp.RLock()
	defer bp.RUnlock()

	if len(bp.mustHitMasks) == 0 {
		utils.LogInst().Debugf("=======>>> empty must hit ip rule is used [%s]\n", srcIP.String())
		return false
	}
	if bp.global {
		return true
	}

	for mip, mask := range bp.mustHitMasks {
		maskIP := srcIP.Mask(mask).String()
		if maskIP == mip {
			utils.LogInst().Debugf("=======>>> must hit success ip:%s->mip:%s mask:%s\n", srcIP, mip, mask.String())
			return true
		}
	}

	return false
}

func (bp *IpRules) LoadMustHits(ips string) {
	bp.mustHitMasks = make(map[string]net.IPMask)

	array := strings.Split(ips, "\n")
	for _, cidr := range array {
		ip, subNet, err := net.ParseCIDR(cidr)
		if err != nil {
			utils.LogInst().Debugf("=======>>> invalid  must hit cidr %s\n", cidr)
			continue
		}
		bp.mustHitMasks[ip.String()] = subNet.Mask
	}
	utils.LogInst().Infof("=======>>> Total must hits :%d \n", len(bp.mustHitMasks))
}

func (bp *IpRules) SetGlobal(g bool) {
	bp.Lock()
	defer bp.Unlock()
	bp.global = g
}

func (bp *IpRules) IsGlobal() bool {
	bp.RLock()
	defer bp.RUnlock()
	return bp.global
}
