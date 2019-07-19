package main

import (
	"encoding/json"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"net"
	"os"
	"runtime/debug"
)

func main() {
	conn, e := net.Dial("tcp", "localhost:8080")

	if e != nil {
		panic(e)
	}
	go Recv(conn)
	var login []byte
	login, e = tcpx.PackWithMarshallerName(tcpx.Message{
		MessageID: 1,
		Header: map[string]interface{}{
			"Authorization": "a client secret key",
		},
		Body: map[string]interface{}{
			"username": "tcpx",
			"password": "123",
		},
	}, "json")

	_, e = conn.Write(login)
	select {}
}

func Recv(conn net.Conn) {
	var buf []byte
	var e error
	for {
		buf, e = tcpx.FirstBlockOf(conn)
		if e != nil {
			handleError(e)
			return
		}
		messageID, e := tcpx.MessageIDOf(buf)
		if e != nil {
			handleError(e)
			return
		}
		body,e := tcpx.BodyBytesOf(buf)
		if e != nil {
			handleError(e)
			return
		}
		switch messageID {
		case 500, 400, 403:
			var m map[string]interface{}
		    e := json.Unmarshal(body, &m)
			if e != nil {
				handleError(e)
				return
			}
			wellPrint(m)
		case 2:
			type Response struct {
				Token string `json:"token"`
			}
			var resp Response
			e := json.Unmarshal(body, &resp)
			if e != nil {
				handleError(e)
				return
			}
			wellPrint(resp)
		}
	}
}

func handleError(e error) {
	if e!=nil {
		fmt.Printf("%v \n %s", e, debug.Stack())
		os.Exit(1)
	}
}

func wellPrint(src interface{}) {
	buf, e := json.MarshalIndent(src, "  ", "  ")
	if e != nil {
		handleError(e)
	}
	fmt.Println(string(buf))
}
