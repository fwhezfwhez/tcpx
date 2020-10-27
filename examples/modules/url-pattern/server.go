package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"tcpx"
)

func main() {
	srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

	srv.Any("/login/", login)

	fmt.Println("tcp listens on 7071")
	if e := srv.ListenAndServe("tcp", ":7071"); e != nil {
		panic(e)
	}
}

func login(c *tcpx.Context) {
	type User struct {
		Username string `json:"username"`
	}

	var user User
	if _, e := c.Bind(&user); e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}
	fmt.Printf("hello %s\n", user.Username)

}
