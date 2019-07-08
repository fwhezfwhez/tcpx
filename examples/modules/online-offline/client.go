package main

import (
	"fmt"
	"os"

	"github.com/fwhezfwhez/tcpx"
	"net"
	//"tcpx"
	"time"
)

func main() {
	conn, e := net.Dial("tcp", "localhost:8102")

	if e != nil {
		panic(e)
	}

	go Recv(conn)
	onlineBuf, e := tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: 1,
		Header:    nil,
		Body:      struct{ Username string `json:"username"` }{"ft"},
	}, tcpx.JsonMarshaller{})

	offlineBuf := tcpx.PackStuff(3)

	conn.Write(onlineBuf)
	time.Sleep(5 * time.Second)
	conn.Write(offlineBuf)

	select {}
}

func Recv(conn net.Conn) {
	var buf = make([]byte, 500)
	for {
		n, e := conn.Read(buf)
		if e != nil {
			os.Exit(0)
			break
		}
		fmt.Println(buf[:n])
		bf, _ := tcpx.BodyBytesOf(buf[:n])
		fmt.Println(string(bf))
	}
}
