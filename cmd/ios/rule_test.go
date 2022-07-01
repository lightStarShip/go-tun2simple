package tun2Simple

import (
	"flag"
	"fmt"
	"github.com/lightStarShip/go-tun2simple/utils"
	"io/ioutil"
	"regexp"
	"testing"
)

var uid = ""

func init() {
	flag.StringVar(&uid, "uid", "", "")
}
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

func TestRuleMatch(t *testing.T) {
	bts, err := ioutil.ReadFile("rule.txt")
	if err != nil {
		t.Fatal(err)
	}
	RInst().Setup(string(bts))
	if RInst().isMatched(uid) {
		utils.LogInst().Infof("======>>>domain[%s] matched", uid)
	}
}

func TestMatchOne1(t *testing.T) {
	bts, err := ioutil.ReadFile("rule.txt")
	if err != nil {
		t.Fatal(err)
	}
	ret := parseRule(string(bts))
	for _, re := range ret {
		if re.MatchString(uid) {
			fmt.Println("bingo:=>", re.String())
			return
		}
		if re.String() == "\\.googleapis.com\\." {
			fmt.Println("matching rex:->", re.String())
			return
		}
	}
	fmt.Println("no match", uid)
}
func TestMatchOne2(t *testing.T) {

	re, err := regexp.Compile("\\.googleapis.com\\.")
	if err != nil {
		t.Fatal(err)
	}
	if re.MatchString(uid) {
		fmt.Println("matching:->", re.String())
		return
	}
	fmt.Println("no match", uid)
}
