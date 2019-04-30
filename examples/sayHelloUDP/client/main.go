package main

import (
	"encoding/json"
	"fmt"
	"github.com/fwhezfwhez/errorx"

	"github.com/fwhezfwhez/tcpx"
	"net"
)

var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})

func main() {
	conn, err := net.Dial("udp", "localhost:7172")
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

	var buf []byte
	var e error
	buf, e = packx.Pack(5, "hello,I am client xiao ming", map[string]interface{}{
		"api": "/tcpx/client1/",
	})
	if e != nil {
		panic(e)
	}
	conn.Write(buf)

	//buf, e = packx.Pack(7, struct {
	//	Username string `json:"username"`
	//}{"FT"}, map[string]interface{}{
	//	"api": "/tcpx/client1/",
	//})
	//if e != nil {
	//	panic(e)
	//}
	//
	//conn.Write(buf)

	//buf, e = packx.Pack(9, struct {
	//	ServiceName string `json:"service_name"`
	//}{"FT"}, map[string]interface{}{
	//	"api": "/tcpx/client1/",
	//})
	//
	//if e != nil {
	//	panic(e)
	//}
	//conn.Write(buf)

	select {}
}

func Receive(conn net.Conn) <-chan []byte {
	var received = make(chan []byte, 200)
	var buffer = make([]byte, 5000, 5000)
	go func() {
		for {
			// udpConn can't Read fixed small size, can only read for all
			n, e := conn.Read(buffer)
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				break
			}
			received <- buffer[0:n]
			continue
		}
	}()
	return received
}

func smartPrint(src interface{}) string {
	buf, _ := json.MarshalIndent(src, "  ", "  ")
	return string(buf)
}
