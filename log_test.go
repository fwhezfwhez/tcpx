package tcpx

import (
	"log"
	"os"
	"testing"
)

var logger *Log

func InitLog() {
	logger = &Log{
		Logger: log.New(os.Stderr, "[tcpx] ", log.LstdFlags|log.Llongfile),
		Mode:   DEBUG,
	}
}
func TestLog_Println(t *testing.T) {
	InitLog()

    logger.Println("test-case logger hello")
    logger.Mode = RELEASE
    logger.Println("test-case hello")

    logger.SetLogMode(DEBUG)
    logger.SetLogFlags(log.Llongfile)

    SetLogFlags(log.Llongfile|log.LUTC)
    SetLogMode(DEBUG)
    Logger.Println("global logger hello")
}

func TestPrintDepth(t *testing.T) {
	InitLog()

	logger.Println("test-case logger hello")
}
