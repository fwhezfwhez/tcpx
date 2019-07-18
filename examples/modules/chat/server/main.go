package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	//"github.com/fwhezfwhez/tcpx"
	"tcpx"
)

func main() {
	srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

	srv.WithBuiltInPool(true)

	srv.AddHandler(1, online)
	srv.AddHandler(3, offline)
	srv.AddHandler(5, send)
	srv.ListenAndServe("tcp", ":8103")
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
}

func offline(c *tcpx.Context) {
	fmt.Println("offline success")
	c.Offline()
}

func send(c *tcpx.Context) {
	type RequestFrom struct {
		Message string `json:"message"`
		ToUser  string `json:"to_user"`
	}
	type ResponseTo struct {
		Message  string `json:"message"`
		FromUser string `json:"from_user"`
	}
	var req RequestFrom
	if _, e := c.Bind(&req); e != nil {
		panic(e)
	}
	if e := c.SendToUsername(req.ToUser, 6, ResponseTo{
		Message:  req.Message,
		FromUser: req.ToUser,
	}); e != nil {
		panic(e)
	}
}
