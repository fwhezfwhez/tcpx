package main

import (
	"fmt"

	"github.com/fwhezfwhez/tcpx"
)

func main() {
	srv := tcpx.NewTcpX(nil)
	srv.OnMessage = func(c *tcpx.Context) {
		var message []byte
		c.Bind(&message)
		fmt.Println(string(message))
	}
	srv.ListenAndServe("tcp", "localhost:8080")
}
