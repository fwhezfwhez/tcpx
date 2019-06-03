package tcpx

import (
	"encoding/json"
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

func TestJSONEmptyMap(t *testing.T) {
	var a = make(map[string]string)
	b,_:=json.Marshal(a)
	fmt.Println(string(b))
}

func TestMessage_Get(t *testing.T) {
	var m = Message{
		MessageID:1,
		Header: map[string]interface{}{"api":"/tcpx/test/"},
		Body: "hello",
	}
	fmt.Println(m.Get("api"))
	if m.Get("api") != "/tcpx/test/" {
		fmt.Println(fmt.Sprintf("key 'api' wanted value '/tcpx/test/' but got '%s'", m.Get("api")))
		t.Fail()
		return
	}
}

func TestMessage_Set(t *testing.T) {
	var m = Message{
		MessageID:1,
		Header: map[string]interface{}{"api":"/tcpx/test/"},
		Body: "hello",
	}
	if m.Set("api", "/tcpx/test/2/"); m.Get("api")!= "/tcpx/test/2/" {
		fmt.Println(fmt.Sprintf("key 'api' wanted value '/tcpx/test/2/' but got '%s'", m.Get("api")))
		t.Fail()
		return
	}

	var m2 Message
	fmt.Println(m2.Get("example_nil"))
}
