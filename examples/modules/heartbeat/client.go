package main

import (
	"github.com/fwhezfwhez/tcpx"
	"net"
	//"tcpx"
	"time"
)

func main() {
	conn, e := net.Dial("tcp", "localhost:8101")

	if e != nil {
		panic(e)
	}
	var heartBeat []byte
	heartBeat, e = tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: tcpx.DEFAULT_HEARTBEAT_MESSAGEID,
		Header:    nil,
		Body:      nil,
	}, nil)
	for {
		conn.Write(heartBeat)
		time.Sleep(10 * time.Second)
	}
}
