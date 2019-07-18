package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	//"tcpx"
)

func main() {
	srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

	srv.WithBuiltInPool(true)

	srv.AddHandler(1, online)
	srv.AddHandler(3, offline)
	srv.ListenAndServe("tcp", ":8102")
}

func online(c *tcpx.Context) {
	type Login struct {
		Username string `json:"username"`
	}
	var login Login
	if _, e := c.Bind(&login); e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}
	c.Online(login.Username)
	fmt.Println("online success")
}

func offline(c *tcpx.Context) {
	fmt.Println("offline success")
	c.Offline()
}
