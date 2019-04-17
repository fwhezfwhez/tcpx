package tcpx

import (
	"fmt"
	"testing"
)

var tcpx = TCPx{Marshaller: JsonMarshaller{}}

func TestTCPx_Pack_UnPack(t *testing.T) {
	type Request struct {
		Username string `json:"username"`
		Age      int    `json:"age"`
	}
	var clientRequest = Request{
		Username: "tcpx",
		Age:      24,
	}
	buf, e :=tcpx.Pack(1, clientRequest, map[string]interface{}{
		"note": "this is a map note",
	})
	if e!=nil {
		panic(e)
	}
	fmt.Println("客户端发送请求:", clientRequest)
	fmt.Println("内容:",buf)

	var serverRequest  Request
	message, e:=tcpx.Unpack(buf, &serverRequest)
	if e !=nil {
		panic(e)
	}
	fmt.Println("收到客户端请求:", serverRequest)
	fmt.Println("客户端信息:", message)
}
