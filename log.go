package tcpx

import (
	"log"
	"os"
	"sync"
)

const (
	DEBUG = 1 + iota
	RELEASE
)

type Log struct {
	Logger *log.Logger
	Mode   int
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
var m sync.RWMutex

func SetLogMode(mode int) {
	Logger.Mode = mode
}
func SetLogFlags(flags int) {
	Logger.Logger.SetFlags(flags)
}
