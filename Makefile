BINDIR=bin

#.PHONY: pbs

all: a i test
#
#pbs:
#	cd pbs/ && $(MAKE)
#

tp:=./

test:
	go build  -ldflags '-w -s' -o $(BINDIR)/ctest mac/*.go

a:
	gomobile bind -v -o $(BINDIR)/tun2Simple.aar -target=android/arm,android/arm64 -ldflags=-s github.com/lightStarShip/go-tun2simple/cmd/android
i:
	gomobile bind -v -o $(BINDIR)/tun2Simple.xcframework -target=ios -ldflags="-s -w" github.com/lightStarShip/go-tun2simple/cmd/ios

clean:
	gomobile clean
	rm $(BINDIR)/*
