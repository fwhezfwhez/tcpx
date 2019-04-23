package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
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

var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})

type Param struct {
	Stream      []byte `json:"stream"`
	MarshalName string `json:"marshal_name"`
}

func main() {
    // TestGoJSON()
}

func TestGoJSON() {
	var c http.Client
	// json
	var userJson = JSONUser{
		Username: "tcpx",
	}
	buf, e := packx.Pack(1, userJson)
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
	req, e := http.NewRequest("POST", "http://localhost:7000/tcpx/clients/stream/", bytes.NewReader(send))
	req.Header.Set("Content-Type","application/json")
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

func Debug(src interface{}) string {
	buf, _ := json.MarshalIndent(src, "  ", "  ")
	return string(buf)
}
