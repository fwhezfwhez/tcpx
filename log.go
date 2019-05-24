package tcpx

import (
	"log"
	"os"
)

const (
	// debug mode, logger of tcpx will print
	DEBUG = 1 + iota
	// release mode, logger of tcpx will not print
	RELEASE
)

// tcpx logger
type Log struct {
	Logger *log.Logger
	Mode   int
}

// Set mode of logger, value is tcpx.DEBUG, tcpx.RELEASE
func (l *Log) SetLogMode(mode int) {
	l.Mode = mode
}

// Set logger flags, value of flags are the same as the official log
func (l *Log) SetLogFlags(flags int) {
	l.Logger.SetFlags(flags)
}

// Println info in debug mode, do nothing in release mode
func (l Log) Println(info ...interface{}) {
	if l.Mode == DEBUG {
		l.Logger.Println(info ...)
	}
}

// Global instance of logger
var Logger = Log{
	Logger: log.New(os.Stderr, "[tcpx] ", log.LstdFlags|log.Llongfile),
	Mode:   DEBUG,
}

// Set global instance logger mode
func SetLogMode(mode int) {
	Logger.Mode = mode
}
// Set global instance logger flags
func SetLogFlags(flags int) {
	Logger.Logger.SetFlags(flags)
}
