package core

import "sync"

type LWIPStack interface {
	Write([]byte) (int, error)
	Close() error
	RestartTimeouts()
}

var (
	_once sync.Once
	_inst LWIPStack
)

func Inst() LWIPStack {
	_once.Do(func() {
		_inst = newLwip()
	})
	return _inst
}
