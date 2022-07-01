package tun2Simple

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestRuleLoad(t *testing.T) {
	bts, err := ioutil.ReadFile("rule.txt")
	if err != nil {
		t.Fatal(err)
	}
	ret := parseRule(string(bts))
	for s, item := range ret {
		fmt.Println(s, item.String())
	}
}
