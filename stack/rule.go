package stack

import (
	"context"
	"github.com/lightStarShip/go-tun2simple/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"regexp"
	"strings"
	"sync"
)

const (
	MaxDnsQueryCnt = 1 << 11
)

var (
	_rOnce sync.Once
	_rInst *Rule
)

type IPCache map[string]string
type Regexps []string

func (ic *IPCache) String() string {
	s := "\n-----ip list------"
	for ip := range *ic {
		s += "\n" + ip
	}
	s += "\n------------------"
	return s
}

type Rule struct {
	msgChan    chan *dnsmessage.Message
	matcher    Regexps
	ipToDomain IPCache
	ipLocker   sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func RInst() *Rule {
	_rOnce.Do(func() {
		_rInst = newRule()
	})

	return _rInst
}

func newRule() *Rule {
	ctx, c := context.WithCancel(context.Background())
	r := &Rule{
		msgChan:    make(chan *dnsmessage.Message, MaxDnsQueryCnt),
		ipToDomain: make(IPCache),
		ctx:        ctx,
		cancel:     c,
	}
	go r.dnsProc()
	return r
}

func (r *Rule) IsMatched(s string) bool {
	for _, re := range r.matcher {
		if ok, err := regexp.MatchString(re, s); ok && err == nil {
			utils.LogInst().Infof("======>>>******matched by [%s] ", re)
			return true
		}
	}
	return false
}

func (r *Rule) Close() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Rule) NeedProxy(ip string) string {
	r.ipLocker.RLock()
	defer r.ipLocker.RUnlock()
	s, ok := r.ipToDomain[ip]
	if !ok {
		return ""
	}
	return s
}

func (r *Rule) dnsProc() {
	utils.LogInst().Infof("======>>> rule manager start to work")
	for {
		select {
		case <-r.ctx.Done():
			utils.LogInst().Infof("======>>> rule manager exit by controller")
			return
		case msg := <-r.msgChan:
			utils.LogInst().Debugf("======>>>dns[%d] answers:%v :=>", msg.ID, msg.Answers)
			var matchedDomain = ""
			for i, question := range msg.Questions {
				domain, typ := question.Name.String(), question.Type.String()
				utils.LogInst().Debugf("======>>>dns[%d] question[%d]:%s typ:%s",
					msg.ID, i, domain, typ)

				if r.IsMatched(domain) {
					matchedDomain = domain
					utils.LogInst().Infof("======>>>[%d]******domain[%s]******matched", msg.ID, domain)
					break
				} else {
					utils.LogInst().Infof("======>>>[%d]++++++domain[%s] ++++++not matched", msg.ID, domain)
				}
			}

			if len(matchedDomain) == 0 {
				utils.LogInst().Infof("======>>>this domain no need to process:%v", msg.Questions)
				continue
			}

			for _, answer := range msg.Answers {
				ar, ok := answer.Body.(*dnsmessage.AResource)
				if !ok {
					utils.LogInst().Warnf("======>>>not ipv4 answer typ:%s", answer.Body.GoString())
					continue
				}
				ip := net.IPv4(ar.A[0], ar.A[1], ar.A[2], ar.A[3]).String()
				utils.LogInst().Infof("======>>>>******[%d]new ip[%s] cached:", msg.ID, ip)
				r.ipLocker.Lock()
				r.ipToDomain[ip] = matchedDomain
				r.ipLocker.Unlock()
			}
		}
	}
}

func (r *Rule) Setup(s string) {
	r.matcher = parseRule(s)
}

func (r *Rule) DirectIPAndHOst(host, ip string) {
	r.ipLocker.RLock()
	if _, ok := r.ipToDomain[ip]; ok {
		r.ipLocker.RUnlock()
		utils.LogInst().Infof("======>>>DirectIPAndHOst [%s]******domain[%s]******cached", ip, host)
		return
	}
	r.ipLocker.RUnlock()
	if !r.IsMatched(host) {
		utils.LogInst().Infof("======>>> DirectIPAndHOst [%s]++++++domain[%s] ++++++not matched", ip, host)
		return
	}
	utils.LogInst().Infof("======>>>DirectIPAndHOst [%s]******domain[%s]******matched", ip, host)
	r.ipLocker.Lock()
	r.ipToDomain[ip] = host
	r.ipLocker.Unlock()
}

func (r *Rule) ParseDns(msg *dnsmessage.Message) {
	utils.LogInst().Debugf("======>>>dns[%d] answers:%v :=>", msg.ID, msg.Answers)
	var matchedDomain = ""
	for i, question := range msg.Questions {
		domain, typ := question.Name.String(), question.Type.String()
		utils.LogInst().Debugf("======>>>dns[%d] question[%d]:%s typ:%s",
			msg.ID, i, domain, typ)

		if r.IsMatched(domain) {
			matchedDomain = domain
			utils.LogInst().Infof("======>>>[%d]******domain[%s]******matched", msg.ID, domain)
			break
		} else {
			utils.LogInst().Infof("======>>>[%d]++++++domain[%s] ++++++not matched", msg.ID, domain)
		}
	}

	if len(matchedDomain) == 0 {
		utils.LogInst().Infof("======>>>this domain no need to process:%v", msg.Questions)
		return
	}

	for _, answer := range msg.Answers {
		ar, ok := answer.Body.(*dnsmessage.AResource)
		if !ok {
			utils.LogInst().Warnf("======>>>not ipv4 answer typ:%s", answer.Body.GoString())
			continue
		}
		ip := net.IPv4(ar.A[0], ar.A[1], ar.A[2], ar.A[3]).String()
		utils.LogInst().Infof("======>>>>******[%d]new ip[%s] cached:", msg.ID, ip)
		r.ipLocker.Lock()
		r.ipToDomain[ip] = matchedDomain
		r.ipLocker.Unlock()
	}
}

func parseRule(s string) Regexps {
	m := make(Regexps, 0)
	domains := strings.Split(s, "\n")
	for _, domain := range domains {
		if len(domain) < 4 {
			continue
		}
		m = append(m, domain)
	}
	utils.LogInst().Infof("======>>> setup rule size:%d\n", len(m))
	return m
}
