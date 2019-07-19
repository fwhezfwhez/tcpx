package main

import (
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	//"tcpx"
)

func main() {
	srv := tcpx.NewTcpX(nil)
	srv.AddHandler(1, beforeLogin, login)

	fmt.Println("tcp listen on :8080")
	if e := srv.ListenAndServe("tcp", ":8080"); e != nil {
		panic(e)
	}
}

func beforeLogin(c *tcpx.Context) {
	header, e := tcpx.HeaderOf(c.Stream)
	if e != nil {
		c.JSON(tcpx.SERVER_ERROR, tcpx.H{
			"message": e.Error(),
		})
		c.Abort()
		return
	}

	auth, ok := header["Authorization"]
	if !ok {
		c.JSON(tcpx.NOT_AUTH, tcpx.H{
			"message": "not found auth in header['Authorization']",
		})
		c.Abort()
		return
	}

	if auth != "a client secret key" {
		c.JSON(tcpx.NOT_AUTH, tcpx.H{
			"message": "auth fail, invalid client secret key",
		})
		c.Abort()
		return
	}
	c.Next()

	fmt.Println("after login, do log")
}
func login(c *tcpx.Context) {
	type User struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var user User
	if _, e := c.Bind(&user); e != nil {
		c.JSON(2, tcpx.H{
			"message": e.Error(),
		})
		return
	}
	fmt.Println("login success")
	token := "AD2KQ33M3ZI56EJ127AI5DK4EO31Q6QWE8ZK75D43ADF412G"
	c.JSON(2, tcpx.H{"message": "login success", "token": token})
}
