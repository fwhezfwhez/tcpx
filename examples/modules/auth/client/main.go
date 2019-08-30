package main

import (
	"crypto/md5"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"net"
	"os"
	"strings"
)

//var appSecret = "hello"

func main() {
	InvalidClient()
	// ValidClient()
	// ClientAuthTimeout()
}

// auth pass, no special tips
func ValidClient() {
	var appSecret = "hello2222"
	conn, e := net.Dial("tcp", "localhost:8104")

	if e != nil {
		panic(e)
	}

	go Recv(conn)

	var auth Auth
	auth.F1 = "ft"
	auth.F2 = "fj"
	auth.Signature = Encrypt(auth, appSecret)

	var authBuf []byte
	authBuf, e = tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: tcpx.DEFAULT_AUTH_MESSAGEID,
		Header:    nil,
		Body:      auth,
	}, nil)

	_ = authBuf
	//_, e = conn.Write(authBuf)
	//if e != nil {
	//	fmt.Println(e.Error())
	//	return
	//}
	select {}
}

// invalid client auth. AppSecret should be 'hello' in server side.But client use 'hello222'.
// Connection is immediately closed by server side and receive message 'signature error'
func InvalidClient() {
	var appSecret = "hello2222"
	conn, e := net.Dial("tcp", "localhost:8104")

	if e != nil {
		panic(e)
	}

	go Recv(conn)

	var auth Auth
	auth.F1 = "ft"
	auth.F2 = "fj"
	auth.Signature = Encrypt(auth, appSecret)

	var authBuf []byte
	authBuf, e = tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: tcpx.DEFAULT_AUTH_MESSAGEID,
		Header:    nil,
		Body:      auth,
	}, nil)

	_, e = conn.Write(authBuf)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	select {}
}

// Connect to server but do nothing.
// After 30s, server auto stop the connection and client receive EOF.
func ClientAuthTimeout() {
	var appSecret = "hello2222"
	conn, e := net.Dial("tcp", "localhost:8104")

	if e != nil {
		panic(e)
	}

	go Recv(conn)

	var auth Auth
	auth.F1 = "ft"
	auth.F2 = "fj"
	auth.Signature = Encrypt(auth, appSecret)

	var authBuf []byte
	authBuf, e = tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: tcpx.DEFAULT_AUTH_MESSAGEID,
		Header:    nil,
		Body:      auth,
	}, nil)

	_ = authBuf
	//_, e = conn.Write(authBuf)
	//if e != nil {
	//	fmt.Println(e.Error())
	//	return
	//}
	select {}
}

func Recv(conn net.Conn) {
	for {
		var buf = make([]byte, 500)
		n, e := conn.Read(buf)
		if e != nil {
			fmt.Println(e.Error())
			os.Exit(1)
		}
		fmt.Println(string(buf[:n]))
	}
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
