package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"net/http"
	"tcpx/examples/use-gateway/pb"
)

// This example shows how to generate expected tcpx stream using official gateway.
// You can get gateway program via package github.com/fwhezfwhez/tcpx/gateway/pack-transfer.
// MarshalName ranges in json,xml,toml,yaml,protobuf
func main() {
	//// valid
	//ExampleJSON()

	//// valid
	//ExampleXML()

	// valid
	ExampleTOML()

	//// valid
	//ExampleYAML()

	//// invalid
	// ExampleProtobuf()
}

func ExampleJSON() {
	// pack example
	type ServiceContent struct {
		Username string `json:"username"`
	}
	type PackRequest struct {
		MarshalName string                 `json:"marshal_name" binding:"required"`
		Stream      []byte                 `json:"stream" binding:"required"`
		MessageID   int32                  `json:"message_id" binding:"required"`
		Header      map[string]interface{} `json:"header"`
	}
	// marshal way range in xml,json,toml,yaml,protobuf, here examples json
	buf, e := json.Marshal(ServiceContent{"hello, tcpx"})
	if e != nil {
		panic(e)
	}
	var packRequest = PackRequest{
		MarshalName: "json",
		Stream:      buf,
		MessageID:   1,
		Header:      map[string]interface{}{"api": "/pack/"},
	}
	reqBuf, e := json.Marshal(packRequest)
	if e != nil {
		panic(e)
	}
	rsp, e := http.Post("http://localhost:7000/gateway/pack/transfer/", "application/json", bytes.NewReader(reqBuf))
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rsBuf, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	type PackResult struct {
		Stream []byte `json:"stream"`
	}
	var packResult PackResult
	e = json.Unmarshal(rsBuf, &packResult)
	if e != nil {
		panic(e)
	}
	fmt.Println(packResult.Stream)

	// unpack example
	type UnPackRequest struct {
		MarshalName string `json:"marshal_name"`
		Stream      []byte `json:"stream"`
	}

	var unPackRequest = UnPackRequest{
		MarshalName: "json",
		Stream:      packResult.Stream,
	}
	reqBuf, e = json.Marshal(unPackRequest)
	if e != nil {
		panic(e)
	}
	rsp, e = http.Post("http://localhost:7000/gateway/unpack/transfer/", "application/json", bytes.NewReader(reqBuf))
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rsBuf, e = ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rsBuf))
	type UnPackResult struct {
		Body ServiceContent `json:"body"`
	}
	unPackResult := UnPackResult{}
	e = json.Unmarshal(rsBuf, &unPackResult)
	if e != nil {
		panic(e)
	}
	fmt.Println(unPackResult.Body)
}

func ExampleTOML() {
	// pack example
	type ServiceContent struct {
		Username string `json:"username" toml:"username"`
	}
	type PackRequest struct {
		MarshalName string                 `json:"marshal_name" binding:"required"`
		Stream      []byte                 `json:"stream" binding:"required"`
		MessageID   int32                  `json:"message_id" binding:"required"`
		Header      map[string]interface{} `json:"header"`
	}
	// marshal way range in xml,json,toml,yaml,protobuf, here examples json
	buf, e := toml.Marshal(ServiceContent{"hello, tcpx"})
	if e != nil {
		panic(e)
	}
	var packRequest = PackRequest{
		MarshalName: "toml",
		Stream:      buf,
		MessageID:   1,
		Header:      map[string]interface{}{"api": "/pack/"},
	}
	reqBuf, e := json.Marshal(packRequest)
	if e != nil {
		panic(e)
	}
	rsp, e := http.Post("http://localhost:7000/gateway/pack/transfer/", "application/json", bytes.NewReader(reqBuf))
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rsBuf, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	type PackResult struct {
		Message string `json:"message"`
		Stream []byte `json:"stream"`
	}
	var packResult PackResult
	e = json.Unmarshal(rsBuf, &packResult)
	if e != nil {
		panic(e)
	}
	fmt.Println(packResult.Stream)

	// unpack example
	type UnPackRequest struct {
		MarshalName string `json:"marshal_name"`
		Stream      []byte `json:"stream"`
	}

	var unPackRequest = UnPackRequest{
		MarshalName: "toml",
		Stream:      packResult.Stream,
	}
	reqBuf, e = json.Marshal(unPackRequest)
	if e != nil {
		panic(e)
	}
	rsp, e = http.Post("http://localhost:7000/gateway/unpack/transfer/", "application/json", bytes.NewReader(reqBuf))
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rsBuf, e = ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rsBuf))
	type UnPackResult struct {
		Body ServiceContent `json:"body"`
	}
	unPackResult := UnPackResult{}
	e = json.Unmarshal(rsBuf, &unPackResult)
	if e != nil {
		panic(e)
	}
	fmt.Println(unPackResult.Body)
}

func ExampleProtobuf() {
	// pack example
	type PackRequest struct {
		MarshalName string                 `json:"marshal_name" binding:"required"`
		Stream      []byte                 `json:"stream" binding:"required"`
		MessageID   int32                  `json:"message_id" binding:"required"`
		Header      map[string]interface{} `json:"header"`
	}
	// marshal way range in xml,json,toml,yaml,protobuf, here examples json
	buf, e := proto.Marshal(&pb.ServiceContent{Username: "hello, tcpx"})
	if e != nil {
		panic(e)
	}
	var packRequest = PackRequest{
		MarshalName: "protobuf",
		Stream:      buf,
		MessageID:   1,
		Header:      map[string]interface{}{"api": "/pack/"},
	}
	reqBuf, e := json.Marshal(packRequest)
	if e != nil {
		panic(e)
	}
	rsp, e := http.Post("http://localhost:7000/gateway/pack/transfer/", "application/json", bytes.NewReader(reqBuf))
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rsBuf, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	type PackResult struct {
		Message string `json:"message"`
		Stream []byte `json:"stream"`
	}
	var packResult PackResult
	e = json.Unmarshal(rsBuf, &packResult)
	if e != nil {
		panic(e)
	}
	fmt.Println(packResult.Message ,packResult.Stream)

	// unpack example
	type UnPackRequest struct {
		MarshalName string `json:"marshal_name"`
		Stream      []byte `json:"stream"`
	}

	var unPackRequest = UnPackRequest{
		MarshalName: "protobuf",
		Stream:      packResult.Stream,
	}
	reqBuf, e = json.Marshal(unPackRequest)
	if e != nil {
		panic(e)
	}
	rsp, e = http.Post("http://localhost:7000/gateway/unpack/transfer/", "application/json", bytes.NewReader(reqBuf))
	if e != nil {
		panic(e)
	}
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	rsBuf, e = ioutil.ReadAll(rsp.Body)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(rsBuf))
	type UnPackResult struct {
		Body pb.ServiceContent `json:"body"`
	}
	unPackResult := UnPackResult{}
	e = json.Unmarshal(rsBuf, &unPackResult)
	if e != nil {
		panic(e)
	}
	fmt.Println(unPackResult.Body)
}
