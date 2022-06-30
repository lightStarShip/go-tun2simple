package utils

import (
	dLog "log"
	"sync"
)

type LogLevel uint8

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type LogOutputFunc func(string2 string, args ...any)

var (
	_inst *Logger = nil
	_once sync.Once
)

type Logger struct {
	level  LogLevel
	opFunc LogOutputFunc
}

func LogInst() *Logger {
	_once.Do(func() {
		_inst = defaultLog()
	})
	return _inst
}

func (l *Logger) InitParam(level LogLevel, of LogOutputFunc) {
	l.level = level
	l.opFunc = of
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) SetOutput(oFunc LogOutputFunc) {
	l.opFunc = oFunc
}
func (l *Logger) Debugf(msg string, args ...interface{}) {
	if l.level <= DEBUG {
		l.output(msg, args...)
	}
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	if l.level <= INFO {
		l.output(msg, args...)
	}
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	if l.level <= WARN {
		l.output(msg, args...)
	}
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	if l.level <= ERROR {
		l.output(msg, args...)
	}
}

func (l *Logger) Fatalf(msg string, args ...interface{}) {
	dLog.Fatalf(msg, args...)
}

func (l *Logger) output(msg string, args ...interface{}) {
	l.opFunc(msg, args...)
}

func defaultLog() *Logger {
	return &Logger{INFO, dLog.Printf}
}
