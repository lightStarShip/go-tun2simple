package core

type LWIPStack interface {
	Write([]byte) (int, error)
	Close() error
	RestartTimeouts()
}

func NewStack() LWIPStack {
	return newLwip()
}
