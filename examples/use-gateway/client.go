package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"github.com/fwhezfwhez/tcpx/examples/use-gateway/pb"
	"github.com/golang/protobuf/proto"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

// This example shows how to generate expected tcpx stream using official gateway.
// You can get gateway program via package github.com/fwhezfwhez/tcpx/gateway/pack-transfer.
// MarshalName ranges in json,xml,toml,yaml,protobuf
func main() {
	//// valid
	ExampleJSON()

	////// valid
	//ExampleXML()
	//
	//// valid
	//ExampleTOML()
	//
	////// valid
	//ExampleYAML()
	//
	////// valid
	//ExampleProtobuf()
}

func ExampleJSON() {
	fmt.Println("marshal name:", "json")

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
	fmt.Println("pack request:")
	fmt.Println(tcpx.Debug(packRequest))
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
	fmt.Println("pack response:")
	fmt.Println(tcpx.Debug(packResult))
	if e != nil {
		panic(e)
	}
	fmt.Println("pack from gateway:")
	fmt.Println(packResult.Stream)

	// unpack example
	type UnPackRequest struct {
		MarshalName string `json:"marshal_name"`
		Stream      []byte `json:"stream"`
	}

	var unPackRequest = UnPackRequest{
		MarshalName: "json",
		// send to message block one time.
		Stream:      append(packResult.Stream, packResult.Stream...),
	}
	reqBuf, e = json.Marshal(unPackRequest)
	if e != nil {
		panic(e)
	}
	fmt.Println("unpack reuest:")
	fmt.Println(tcpx.Debug(unPackRequest))
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

	type Result struct {
		MessageID   int32                  `json:"message_id"`
		Header      map[string]interface{} `json:"header"`
		MarshalName string                 `json:"marshal_name"`
		Stream      []byte                 `json:"stream"`
	}
	type UnpackResponse struct {
		Message string   `json:"message"`
		Blocks  []Result `json:"blocks"`
	}
	var response UnpackResponse
	e = json.Unmarshal(rsBuf, &response)
	if e != nil {
		panic(e)
	}
	fmt.Println("unpack response:")
	fmt.Println(tcpx.Debug(response))

	// we took {"username":"hello, tcpx"} for example, so response.Blocks length =1
	var value ServiceContent

	// If request stream is more than one messages,
	// you should unmarshal response stream like
	// 	e = json.Unmarshal(response.Blocks[0].Stream, &value)
	// 	e = json.Unmarshal(response.Blocks[1].Stream, &value1)
	// 	e = json.Unmarshal(response.Blocks[2].Stream, &value2)
	// rather than json.Unmarshal(response.Block, &values), values refers to []Value
	e = json.Unmarshal(response.Blocks[0].Stream, &value)
	if e != nil {
		panic(e)
	}
	if value.Username != "hello, tcpx" {
		panic("unpack wrong")
	}
	fmt.Println(value)
}

func ExampleTOML() {
	fmt.Println("marshal name:", "toml")

	// pack example
	type ServiceContent struct {
		Username string `toml:"username"`
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
		Stream []byte `json:"stream"`
	}
	var packResult PackResult
	e = json.Unmarshal(rsBuf, &packResult)
	if e != nil {
		panic(e)
	}
	fmt.Println("pack from gateway:")
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
	type Result struct {
		MessageID   int32                  `json:"message_id"`
		Header      map[string]interface{} `json:"header"`
		MarshalName string                 `json:"marshal_name"`
		Stream      []byte                 `json:"stream"`
	}
	type UnpackResponse struct {
		Message string   `json:"message"`
		Blocks  []Result `json:"blocks"`
	}
	var response UnpackResponse
	e = json.Unmarshal(rsBuf, &response)
	if e != nil {
		panic(e)
	}
	fmt.Println("unpack from gateway:")
	fmt.Println(tcpx.Debug(response))

	// we took {"username":"hello, tcpx"} for example, so response.Blocks length =1
	var value ServiceContent
	e = toml.Unmarshal(response.Blocks[0].Stream, &value)
	if e != nil {
		panic(e)
	}
	if value.Username != "hello, tcpx" {
		panic("unpack wrong")
	}
	fmt.Println(value)
}

func ExampleYAML() {
	fmt.Println("marshal name:", "yaml")
	// pack example
	type ServiceContent struct {
		Username string `yaml:"username"`
	}
	type PackRequest struct {
		MarshalName string                 `json:"marshal_name" binding:"required"`
		Stream      []byte                 `json:"stream" binding:"required"`
		MessageID   int32                  `json:"message_id" binding:"required"`
		Header      map[string]interface{} `json:"header"`
	}
	// marshal way range in xml,json,toml,yaml,protobuf, here examples json
	buf, e := yaml.Marshal(ServiceContent{"hello, tcpx"})
	if e != nil {
		panic(e)
	}
	var packRequest = PackRequest{
		MarshalName: "yaml",
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
	fmt.Println("pack from gateway:")
	fmt.Println(packResult.Stream)

	// unpack example
	type UnPackRequest struct {
		MarshalName string `json:"marshal_name"`
		Stream      []byte `json:"stream"`
	}

	var unPackRequest = UnPackRequest{
		MarshalName: "yaml",
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
	type Result struct {
		MessageID   int32                  `json:"message_id"`
		Header      map[string]interface{} `json:"header"`
		MarshalName string                 `json:"marshal_name"`
		Stream      []byte                 `json:"stream"`
	}
	type UnpackResponse struct {
		Message string   `json:"message"`
		Blocks  []Result `json:"blocks"`
	}
	var response UnpackResponse
	e = json.Unmarshal(rsBuf, &response)
	if e != nil {
		panic(e)
	}
	fmt.Println("unpack from gateway:")
	fmt.Println(tcpx.Debug(response))

	// we took {"username":"hello, tcpx"} for example, so response.Blocks length =1
	var value ServiceContent
	e = yaml.Unmarshal(response.Blocks[0].Stream, &value)
	if e != nil {
		panic(e)
	}
	if value.Username != "hello, tcpx" {
		panic("unpack wrong")
	}
	fmt.Println(value)
}

func ExampleXML() {
	fmt.Println("marshal name:", "xml")

	// pack example
	type ServiceContent struct {
		MLName   xml.Name `xml:"xml"`
		Username string   `xml:"username"`
	}
	type PackRequest struct {
		MarshalName string                 `json:"marshal_name" binding:"required"`
		Stream      []byte                 `json:"stream" binding:"required"`
		MessageID   int32                  `json:"message_id" binding:"required"`
		Header      map[string]interface{} `json:"header"`
	}
	// marshal way range in xml,json,toml,yaml,protobuf, here examples json
	buf, e := xml.Marshal(ServiceContent{Username: "hello, tcpx"})
	if e != nil {
		panic(e)
	}
	var packRequest = PackRequest{
		MarshalName: "xml",
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
	fmt.Println("pack from gateway:")
	fmt.Println(packResult.Stream)

	// unpack example
	type UnPackRequest struct {
		MarshalName string `json:"marshal_name"`
		Stream      []byte `json:"stream"`
	}

	var unPackRequest = UnPackRequest{
		MarshalName: "xml",
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
	type Result struct {
		MessageID   int32                  `json:"message_id"`
		Header      map[string]interface{} `json:"header"`
		MarshalName string                 `json:"marshal_name"`
		Stream      []byte                 `json:"stream"`
	}
	type UnpackResponse struct {
		Message string   `json:"message"`
		Blocks  []Result `json:"blocks"`
	}
	var response UnpackResponse
	e = json.Unmarshal(rsBuf, &response)
	if e != nil {
		panic(e)
	}
	fmt.Println("unpack from gateway:")
	fmt.Println(tcpx.Debug(response))

	// we took {"username":"hello, tcpx"} for example, so response.Blocks length =1
	var value ServiceContent
	e = xml.Unmarshal(response.Blocks[0].Stream, &value)
	if e != nil {
		panic(e)
	}
	if value.Username != "hello, tcpx" {
		panic("unpack wrong")
	}
	fmt.Println(value)
}

func ExampleProtobuf() {
	fmt.Println("marshal name:", "protobuf")

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
		Stream []byte `json:"stream"`
	}
	var packResult PackResult
	e = json.Unmarshal(rsBuf, &packResult)
	if e != nil {
		panic(e)
	}
	fmt.Println("pack from gateway:")
	fmt.Println(packResult.Stream)

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
	type Result struct {
		MessageID   int32                  `json:"message_id"`
		Header      map[string]interface{} `json:"header"`
		MarshalName string                 `json:"marshal_name"`
		Stream      []byte                 `json:"stream"`
	}
	type UnpackResponse struct {
		Message string   `json:"message"`
		Blocks  []Result `json:"blocks"`
	}
	var response UnpackResponse
	e = json.Unmarshal(rsBuf, &response)
	if e != nil {
		panic(e)
	}
	fmt.Println("unpack from gateway:")
	fmt.Println(tcpx.Debug(response))

	// we took {"username":"hello, tcpx"} for example, so response.Blocks length =1
	var value pb.ServiceContent
	e = proto.Unmarshal(response.Blocks[0].Stream, &value)
	if e != nil {
		panic(e)
	}
	if value.Username != "hello, tcpx" {
		panic("unpack wrong")
	}
	fmt.Println(value)
}
