package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"net"
	"os"
	"tcpx"
)

func main() {
	message := tcpx.NewURLPatternMessage("/login/", map[string]interface{}{
		"username": "Li Hua",
	})

	b, e := message.Pack(tcpx.JsonMarshaller{})
	if e != nil {
		panic(e)
	}

	conn, e := net.Dial("tcp", "localhost:7071")
	if e != nil {
		panic(e)
		return
	}

	go Recv(conn)
	conn.Write(b)

	select {}

}

func Recv(conn net.Conn) {
	for {
		var buf = make([]byte, 500)
		n, e := conn.Read(buf)
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			os.Exit(1)
		}
		fmt.Println(string(buf[:n]))
	}
}
