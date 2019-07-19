package main

import (
	"fmt"
	"os"

	"github.com/fwhezfwhez/tcpx"
	"net"
)

func main() {
	conn, e := net.Dial("tcp", "localhost:8080")
	if e != nil {
		panic(e)
	}
	go Recv(conn)
	select {}
}

func Recv(conn net.Conn) {
	var buf = make([]byte, 500)
	for {
		n, e := conn.Read(buf)
		if e != nil {
			fmt.Println(e.Error())
			os.Exit(0)
			break
		}
		fmt.Println(buf[:n])
		bf, _ := tcpx.BodyBytesOf(buf[:n])
		fmt.Println(string(bf))
	}
}
