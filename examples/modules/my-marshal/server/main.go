package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	"github.com/fwhezfwhez/tcpx/examples/modules/my-marshal/marshaller"
)

var packx = tcpx.NewPackx(marshaller.ByteMarshaller{})

func main() {
	srv := tcpx.NewTcpX(marshaller.ByteMarshaller{})
	srv.AddHandler(22, func(c *tcpx.Context) {
		var message []byte
		mi, e := c.Bind(&message)
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		fmt.Println(mi.MessageID, string(message))
	})

	srv.AddHandler(23, func(c *tcpx.Context) {
		var message []byte
		mi, e := packx.Unpack(c.Stream, &message)
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		fmt.Println(mi.MessageID, string(message))
	})
	fmt.Println("listen on :7011")
	srv.ListenAndServe("tcp", "localhost:7011")
}
