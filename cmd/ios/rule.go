package tun2Simple

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	_rOnce sync.Once
	_rInst *Rule
)

type DomainCache map[string]*RuleItem
type IPCache map[net.Addr]bool

func (ic *IPCache) String() string {
	s := "\n-----ip list------"
	for addr := range *ic {
		s += "\n" + addr.String()
	}
	s += "\n------------------"
	return s
}

type RuleItem struct {
	domain string
	IPAddr IPCache
}

func (ri *RuleItem) String() string {
	return fmt.Sprintf("\n domain:[%s]=>%s\n", ri.domain, ri.IPAddr.String())
}

type Rule struct {
	Domains DomainCache
}

func RInst() *Rule {
	_rOnce.Do(func() {
		_rInst = newRule()
	})

	return _rInst
}

func newRule() *Rule {
	r := &Rule{
		Domains: make(map[string]*RuleItem),
	}
	return r
}

func (r *Rule) Setup(s string) {
	r.Domains = parseRule(s)
}

func parseRule(s string) DomainCache {
	dc := make(DomainCache)
	domains := strings.Split(s, "\n")
	for _, domain := range domains {
		dc[domain] = &RuleItem{
			domain: domain,
		}
	}
	return dc
}
