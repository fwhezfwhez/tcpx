package main

import (
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	//"tcpx"
)

func main() {
	srv := tcpx.NewTcpX(nil)
	srv.UseGlobal(func(c *tcpx.Context) {
		fmt.Println("before raw message in")
	})
	srv.Use("middle-1", func(c *tcpx.Context) {
		fmt.Println("use middleware 1")
	})
	srv.HandleRaw = func(c *tcpx.Context) {
		var buf = make([]byte, 500)
		var n int
		var e error
		for {
			n, e = c.ConnReader.Read(buf)
			if e != nil {
				fmt.Println(e.Error())
				return
			}
			fmt.Println("receive:", string(buf[:n]))
			c.ConnWriter.Write([]byte("hello,I am server."))
		}
	}
	srv.ListenAndServeRaw("tcp", ":6631")
}
