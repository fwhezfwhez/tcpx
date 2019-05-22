package tcpx

import (
	"log"
	"os"
)

const (
	DEBUG = 1 + iota
	RELEASE
)

type Log struct {
	Logger *log.Logger
	Mode   int
}

func (l *Log) SetLogMode(mode int) {
	l.Mode = mode
}
func (l *Log) SetLogFlags(flags int) {
	l.Logger.SetFlags(flags)
}

func (l Log) Println(info ...interface{}) {
	if l.Mode == DEBUG {
		l.Logger.Println(info ...)
	}
}

var Logger = Log{
	Logger: log.New(os.Stderr, "[tcpx] ", log.LstdFlags|log.Llongfile),
	Mode:   DEBUG,
}

func SetLogMode(mode int) {
	Logger.Mode = mode
}
func SetLogFlags(flags int) {
	Logger.Logger.SetFlags(flags)
}
