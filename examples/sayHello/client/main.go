package main

import (
	"encoding/json"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"net"
	"tcpx"
)

var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})

func main() {
	conn, err := net.Dial("tcp", "localhost:7171")
	if err != nil {
		panic(err)
	}
	received := Receive(conn)
	go func() {
		for {
			buf := <-received
			var message tcpx.Message
			var receivedString string
			fmt.Println(buf)
			message, e := packx.Unpack(buf, &receivedString)
			if e != nil {
				panic(errorx.Wrap(e))
			}
			fmt.Println("收到服务端消息块:", smartPrint(message))
			fmt.Println("服务端消息:", receivedString)
		}
	}()
	buf, e := packx.Pack(5, "hello,I am client xiao ming", map[string]interface{}{
		"api": "/tcpx/client1/",
	})

	if e != nil {
		panic(e)
	}
	conn.Write(buf)
	select {}
}

func Receive(conn net.Conn) <-chan []byte {
	var received = make(chan []byte, 200)
	go func() {
		for {
			buf, e := packx.FirstBlockOf(conn)
			if e != nil {
				fmt.Println(e.Error())
				break
			}
			received <- buf
			continue
		}
	}()
	return received
}

func smartPrint(src interface{}) string {
	buf, _ := json.MarshalIndent(src, "  ", "  ")
	return string(buf)
}
