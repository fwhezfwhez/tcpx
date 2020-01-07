package tcpx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx/examples/sayHello/client/pb"
	"testing"
)

var pack = Packx{Marshaller: JsonMarshaller{}}

func TestTCPx_Pack_UnPack(t *testing.T) {
	type Request struct {
		Username string `json:"username"`
		Age      int    `json:"age"`
	}
	var clientRequest = Request{
		Username: "packx",
		Age:      24,
	}
	buf, e := pack.Pack(1, clientRequest, map[string]interface{}{
		"note": "this is a map note",
	})
	if e != nil {
		panic(e)
	}
	fmt.Println("客户端发送请求:", clientRequest)
	fmt.Println("内容:", buf)

	var serverRequest Request
	message, e := pack.Unpack(buf, &serverRequest)
	if e != nil {
		panic(e)
	}
	fmt.Println("收到客户端请求:", serverRequest)
	fmt.Println("客户端信息:", message)
}

func TestTCPx_Packx_Property(t *testing.T) {
	type Request struct {
		Username string `json:"username"`
		Age      int    `json:"age"`
	}
	var clientRequest = Request{
		Username: "packx",
		Age:      24,
	}
	buf, e := pack.Pack(1, clientRequest, map[string]interface{}{
		"note": "this is a map note",
	})
	if e != nil {
		panic(e)
	}
	fmt.Println("客户端发送请求:", clientRequest)
	fmt.Println("内容:", buf)
	fmt.Println(pack.MessageIDOf(buf))
	fmt.Println(pack.HeaderLengthOf(buf))
	fmt.Println(pack.BodyLengthOf(buf))
	fmt.Println(pack.HeaderBytesOf(buf))
	fmt.Println(pack.BodyBytesOf(buf))
	fmt.Println(packx.HeaderOf(buf))

	header, _ := pack.HeaderBytesOf(buf)

	body, _ := pack.BodyBytesOf(buf)

	var result Request
	e = json.Unmarshal(body, &result)
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	fmt.Println(result)
	var resultHeader map[string]interface{}
	e = json.Unmarshal(header, &resultHeader)
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	fmt.Println(resultHeader)
}

func TestPackx_PackWithBody(t *testing.T) {
	packx := NewPackx(JsonMarshaller{})
	buf, e := packx.PackWithBody(1, newBytes(1, 2, 3, 4, 5))
	if e != nil {
		fmt.Println(errorx.Wrap(e))
		t.Fail()
		return
	}

	body, e := packx.BodyBytesOf(buf)
	if e != nil {
		fmt.Println(errorx.Wrap(e))
		t.Fail()
		return
	}
	fmt.Println(body)
	if len(body) != 5 {
		fmt.Println(fmt.Sprintf("body unpack want length 5 but got %d, %v", len(body), body))
		t.Fail()
		return
	}
}
func TestPackWithMarshallerName_UnPackWithUnmarshalName(t *testing.T) {
	// xml
	{
		buf, e := PackWithMarshallerName(Message{
			MessageID: 1,
			Header:    nil,
			Body: struct {
				XMLName xml.Name `xml:"xml"`
				Name    string   `xml:"name"`
			}{
				Name: "hello",
			},
		}, "xml")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		var receive2 struct {
			XMLName xml.Name `xml:"xml"`
			Name    string   `xml:"name"`
		}
		_, e = UnpackWithMarshallerName(buf, &receive2, "xml")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		if receive2.Name != "hello" {
			fmt.Println(fmt.Sprintf("received want %s but got %s", "hello", receive2))
			t.Fail()
			return
		}
	}

	// json
	{
		buf, e := PackWithMarshallerName(Message{
			MessageID: 1,
			Header:    nil,
			Body:      "hello",
		}, "json")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		var receive string
		_, e = UnpackWithMarshallerName(buf, &receive, "json")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		if receive != "hello" {
			fmt.Println(fmt.Sprintf("received want %s but got %s", "hello", receive))
			t.Fail()
			return
		}
	}

	// toml
	{
		buf, e := PackWithMarshallerName(Message{
			MessageID: 1,
			Header:    nil,
			Body: struct {
				Name string `toml:"name"`
			}{
				Name: "hello",
			},
		}, "toml")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		var receive2 struct {
			Name string `toml:"name"`
		}
		_, e = UnpackWithMarshallerName(buf, &receive2, "toml")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		if receive2.Name != "hello" {
			fmt.Println(fmt.Sprintf("received want %s but got %s", "hello", receive2))
			t.Fail()
			return
		}
	}

	// yaml
	{
		buf, e := PackWithMarshallerName(Message{
			MessageID: 1,
			Header:    nil,
			Body: struct {
				Name string `yaml:"name"`
			}{
				Name: "hello",
			},
		}, "yaml")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		var receive2 struct {
			Name string `yaml:"name"`
		}
		_, e = UnpackWithMarshallerName(buf, &receive2, "yaml")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		if receive2.Name != "hello" {
			fmt.Println(fmt.Sprintf("received want %s but got %s", "hello", receive2))
			t.Fail()
			return
		}
	}

	// protobuf
	{
		obj := pb.SayHelloRequest{
			Username: "hello",
		}
		buf, e := PackWithMarshallerName(Message{
			MessageID: 1,
			Header:    nil,
			Body:      &obj,
		}, "protobuf")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		var receive pb.SayHelloRequest
		_, e = UnpackWithMarshallerName(buf, &receive, "protobuf")
		if e != nil {
			fmt.Println(errorx.Wrap(e))
			t.Fail()
			return
		}
		if receive.Username != "hello" {
			fmt.Println(fmt.Sprintf("received want %s but got %s", "hello", receive.Username))
			t.Fail()
			return
		}
	}
}
func newBytes(a ...byte) []byte {
	return a
}

func TestPack(t *testing.T) {
	PackWithMarshaller(Message{
		MessageID: 1,
		Header: map[string]interface{}{
			"auth": "abc",
		},
		Body: map[string]interface{}{
			"username":"tcpx",
		},
	}, JsonMarshaller{})
	PackWithMarshallerAndBody()
}
