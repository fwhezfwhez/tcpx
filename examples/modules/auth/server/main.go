package main

import (
	"crypto/md5"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"strings"
	"time"
)

var appSecret = "hello"
func main() {
	srv := tcpx.NewTcpX(nil)
	tcpx.SetLogMode(tcpx.DEBUG)
	srv.WithAuthDetail(true, 30*time.Second, false, tcpx.DEFAULT_AUTH_MESSAGEID, func(c *tcpx.Context) {
		var auth Auth

		if _, e := c.Bind(&auth); e != nil {
			c.Conn.Write([]byte("server error: " + e.Error()))
			return
		}
		if Encrypt(auth, appSecret) != auth.Signature {
			c.Conn.Write([]byte("auth deny, signature wrong"))
			c.RecvAuthDeny()
			return
		}
		c.RecvAuthPass()
	})
    fmt.Println("tcp start on :8104")
	srv.ListenAndServe("tcp", ":8104")
}

type Auth struct {
	F1        string
	F2        string
	Signature string
}

func Encrypt(a Auth, secret string) string {
	return tcpx.MD5(a.F1 + a.F2 + secret)
}
func MD5(rawMsg string) string {
	data := []byte(rawMsg)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has)
	return strings.ToUpper(md5str1)
}
