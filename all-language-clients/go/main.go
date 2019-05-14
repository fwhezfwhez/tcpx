package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"github.com/fwhezfwhez/tcpx/all-language-clients/model"
	"io/ioutil"
	"net/http"
)

type JSONUser struct {
	Username string `json:"username"`
}
type XMLUser struct {
	XMLName  xml.Name `xml:"xml"`
	Username string   `xml:"username"`
}
type TOMLUser struct {
	Username string `toml:"username"`
}
type YAMLUser struct {
	Username string `yaml:"username"`
}

type Param struct {
	Stream      []byte `json:"stream"`
	MarshalName string `json:"marshal_name"`
}

func main() {
	//TestGoJSON()
	 TestGoProtoBuf()
	// TestGoTOML()
	// TestGoYAML()
	// TestGoXML()
}

func TestGoProtoBuf() {
	var c http.Client
	var packx = tcpx.NewPackx(tcpx.ProtobufMarshaller{})
	// json
	var userProto = model.User{
		Username: "tcpx",
	}

	buf, e := packx.Pack(1, &userProto)

	if e != nil {
		panic(e)
	}
	var param = Param{
		Stream:      buf,
		MarshalName: "protobuf",
	}
	send, e := json.Marshal(param)
	if e != nil {
		panic(e)
	}
	req, e := http.NewRequest("POST", "http://localhost:7001/tcpx/clients/stream/", bytes.NewReader(send))
	req.Header.Set("Content-Type", "application/json")
	if e != nil {
		panic(e)
	}
	rsp, e := c.Do(req)
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rs, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rs))
}

func TestGoJSON() {
	var c http.Client
	var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})
	// json
	var userJson = JSONUser{
		Username: "tcpx",
	}
	buf, e := packx.Pack(1, userJson)
	fmt.Println(buf)
	if e != nil {
		panic(e)
	}
	var param = Param{
		Stream:      buf,
		MarshalName: "json",
	}
	send, e := json.Marshal(param)
	if e != nil {
		panic(e)
	}
	req, e := http.NewRequest("POST", "http://localhost:7001/tcpx/clients/stream/", bytes.NewReader(send))
	req.Header.Set("Content-Type", "application/json")
	if e != nil {
		panic(e)
	}
	rsp, e := c.Do(req)
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rs, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rs))
}

func TestGoTOML() {
	var c http.Client
	var packx = tcpx.NewPackx(tcpx.TomlMarshaller{})
	// json
	var userToml = TOMLUser{
		Username: "tcpx",
	}
	buf, e := packx.Pack(1, userToml)
	if e != nil {
		panic(e)
	}
	var param = Param{
		Stream:      buf,
		MarshalName: "toml",
	}
	send, e := json.Marshal(param)
	if e != nil {
		panic(e)
	}
	req, e := http.NewRequest("POST", "http://localhost:7001/tcpx/clients/stream/", bytes.NewReader(send))
	req.Header.Set("Content-Type", "application/json")
	if e != nil {
		panic(e)
	}
	rsp, e := c.Do(req)
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rs, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rs))
}

func TestGoYAML() {
	var c http.Client
	var packx = tcpx.NewPackx(tcpx.YamlMarshaller{})
	// json
	var userYaml = YAMLUser{
		Username: "tcpx",
	}
	buf, e := packx.Pack(1, userYaml)
	if e != nil {
		panic(e)
	}
	var param = Param{
		Stream:      buf,
		MarshalName: "yaml",
	}
	send, e := json.Marshal(param)
	if e != nil {
		panic(e)
	}
	req, e := http.NewRequest("POST", "http://localhost:7001/tcpx/clients/stream/", bytes.NewReader(send))
	req.Header.Set("Content-Type", "application/json")
	if e != nil {
		panic(e)
	}
	rsp, e := c.Do(req)
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rs, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rs))
}

func TestGoXML() {
	var c http.Client
	var packx = tcpx.NewPackx(tcpx.XmlMarshaller{})
	// json
	var userXml = XMLUser{
		Username: "tcpx",
	}
	buf, e := packx.Pack(1, userXml)
	if e != nil {
		panic(e)
	}
	var param = Param{
		Stream:      buf,
		MarshalName: "xml",
	}
	fmt.Println(Debug(param))
	send, e := json.Marshal(param)
	if e != nil {
		panic(e)
	}
	req, e := http.NewRequest("POST", "http://localhost:7001/tcpx/clients/stream/", bytes.NewReader(send))
	req.Header.Set("Content-Type", "application/json")
	if e != nil {
		panic(e)
	}
	rsp, e := c.Do(req)
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rs, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rs))
}

func Debug(src interface{}) string {
	buf, _ := json.MarshalIndent(src, "  ", "  ")
	return string(buf)
}
