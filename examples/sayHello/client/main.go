package main

import (
	"encoding/json"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"io"
	"tcpx/examples/sayHello/client/pb"

	"github.com/fwhezfwhez/tcpx"
	"net"
	//"tcpx"
)

var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})
var packxProto = tcpx.NewPackx(tcpx.ProtobufMarshaller{})

func main() {
	conn, err := net.Dial("tcp", "localhost:7171")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			buf ,e :=packxProto.FirstBlockOf(conn)
			if e != nil {
				if e == io.EOF {
					break
				}
				panic(errorx.Wrap(e))
			}
			//var receivedString string
			////fmt.Println(buf)
			//message, e := packx.Unpack(buf, &receivedString)
			//if e != nil {
			//	panic(errorx.Wrap(e))
			//}
			//fmt.Println("收到服务端消息块:", smartPrint(message))
			//fmt.Println("服务端消息:", receivedString)
			var resp pb.SayHelloReponse
			message, e := packxProto.Unpack(buf, &resp)

			fmt.Println("收到服务端消息块:", smartPrint(message))
			fmt.Println("服务端消息:", resp)
		}
	}()

	var buf []byte
	var e error
	buf, e = packxProto.Pack(11, &pb.SayHelloRequest{
		Username: "ft",
	})
	if e != nil {
		panic(e)
	}
	conn.Write(buf)

	//buf, e = packx.Pack(5, "hello,I am client xiao ming", map[string]interface{}{
	//	"api": "/tcpx/client1/",
	//})
	//if e != nil {
	//	panic(e)
	//}
	//
	//fmt.Println(buf)
	//conn.Write(buf)

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


func smartPrint(src interface{}) string {
	buf, _ := json.MarshalIndent(src, "  ", "  ")
	return string(buf)
}
