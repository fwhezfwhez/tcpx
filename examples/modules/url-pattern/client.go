package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"net"
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
		block, e := tcpx.Recv(conn)
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		urlPattern, e := block.URLPattern()
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}

		switch urlPattern {
		case "/login/":
			type LoginResponse struct {
				Token string `json:"token"`
			}
			var lr LoginResponse
			block.BindJSON(&lr)
			fmt.Printf("recv login token: %s\n", lr.Token)
		default:
			fmt.Println("unexpected urlPattern")
		}

	}
}
