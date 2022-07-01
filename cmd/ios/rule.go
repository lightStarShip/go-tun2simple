package tun2Simple

import (
	"github.com/lightStarShip/go-tun2simple/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"regexp"
	"strings"
	"sync"
)

var (
	_rOnce sync.Once
	_rInst *Rule
)

type IPCache map[string]bool
type Regexps []*regexp.Regexp

func (ic *IPCache) String() string {
	s := "\n-----ip list------"
	for ip := range *ic {
		s += "\n" + ip
	}
	s += "\n------------------"
	return s
}

type Rule struct {
	msgChan chan *dnsmessage.Message
	matcher Regexps
	ips     IPCache
}

func RInst() *Rule {
	_rOnce.Do(func() {
		_rInst = newRule()
	})

	return _rInst
}

func newRule() *Rule {
	r := &Rule{
		msgChan: make(chan *dnsmessage.Message, 1024),
		ips:     make(IPCache),
	}
	go r.dnsProc()
	return r
}

func (r *Rule) isMatched(s string) bool {
	for _, re := range r.matcher {
		if re.MatchString(s) {
			utils.LogInst().Infof("======>>>******matched by [%s] ", re.String())
			return true
		}
	}
	return false
}

func (r *Rule) NeedProxy(ip string) bool {
	return r.ips[ip]
}

func (r *Rule) dnsProc() {

	utils.LogInst().Infof("======>>> rule manager start to work")
	for {
		select {
		case msg := <-r.msgChan:
			utils.LogInst().Debugf("======>>>dns[%d] answers:%v :=>", msg.ID, msg.Answers)
			var needProcess = false
			for i, question := range msg.Questions {
				domain, typ := question.Name.String(), question.Type.String()
				utils.LogInst().Debugf("======>>>dns[%d] question[%d]:%s typ:%s",
					msg.ID, i, domain, typ)

				if r.isMatched(domain) {
					needProcess = true
					utils.LogInst().Infof("======>>>[%d]******domain[%s] matched", msg.ID, domain)
				} else {
					utils.LogInst().Infof("======>>>[%d]++++++domain[%s] not matched", msg.ID, domain)
				}
			}

			if !needProcess {
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
				r.ips[ip] = true
			}
		}
	}
}

func (r *Rule) Setup(s string) {
	r.matcher = parseRule(s)
}

func (r *Rule) ParseDns(msg *dnsmessage.Message) {
	r.msgChan <- msg
}

func parseRule(s string) Regexps {
	m := make(Regexps, 0)
	domains := strings.Split(s, "\n")
	for _, domain := range domains {
		if len(domain) < 4 {
			continue
		}
		re, err := regexp.Compile(domain)
		if err != nil {
			utils.LogInst().Errorf("======>>> rule[%s] compile err:%v", domain, err)
			continue
		}
		m = append(m, re)
	}
	utils.LogInst().Infof("======>>> setup rule size:%d\n", len(m))
	return m
}
