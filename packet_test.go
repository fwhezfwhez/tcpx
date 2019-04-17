package tcpx

import (
	"fmt"
	"testing"
)

func TestMessage_Packet(t *testing.T) {
	message := Message{
		MessageID: 1,
		Header: nil,
		Body: "你好",
	}
	packet,e := PackWithMarshaller(message, nil)
	if e!=nil {
		panic(e)
	}
	fmt.Println(packet)
	fmt.Println(string(packet))
}

func TestMessage_Unpack(t *testing.T) {
	message := Message{
		MessageID: 1,
		Header: nil,
		Body: "你好",
	}
	packet,e := PackWithMarshaller(message, nil)
	if e!=nil {
		panic(e)
	}
	var body string
	message, e = UnpackWithMarshaller(packet, &body, nil)
	if e!=nil {
		panic(e)
	}
	fmt.Println("message:",message)
	fmt.Println("body:", body)
}
