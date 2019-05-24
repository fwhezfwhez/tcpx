// Package docker-image executable file
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	"github.com/fwhezfwhez/tcpx/all-language-clients/model"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"io"
	"net/http"
	"time"

	//"tcpx"
)

func main() {
	// tcp,udp,kcp on 7171, 7172, 7173
	go Service()
	// gateway 7000
	go GateWay()
	// validation 7001
	go Validate()
	select {}
}

func GateWay() {
	r := gin.Default()

	// 发送包按照每条消息
	r.POST("/gateway/pack/transfer/", func(c *gin.Context) {
		type Param struct {
			MarshalName string                 `json:"marshal_name" binding:"required"`
			Stream      []byte                 `json:"stream" binding:"required"`
			MessageID   int32                  `json:"message_id" binding:"required"`
			Header      map[string]interface{} `json:"header"`
		}
		var param Param
		if e := c.Bind(&param); e != nil {
			c.JSON(400, gin.H{"message": errorx.Wrap(e).Error()})
			return
		}
		fmt.Println(param.Stream)
		marshaller, e := tcpx.GetMarshallerByMarshalName(param.MarshalName)
		fmt.Println(marshaller.MarshalName())
		if e != nil {
			c.JSON(400, gin.H{"message": errorx.Wrap(e).Error()})
			return
		}
		var packx = tcpx.NewPackx(marshaller)
		buf, e := packx.PackWithBody(param.MessageID, param.Stream, param.Header)
		if e != nil {
			c.JSON(400, gin.H{"message": errorx.Wrap(e).Error()})
			return
		}

		c.JSON(200, gin.H{"message": "success", "stream": buf})
	})

	// 接受包会自动分块
	r.POST("/gateway/unpack/transfer/", func(c *gin.Context) {
		type Param struct {
			MarshalName string `json:"marshal_name"`
			Stream      []byte `json:"stream"`
		}
		var param Param
		c.Bind(&param)
		var reader = bytes.NewReader(param.Stream)
		marshaller, e := tcpx.GetMarshallerByMarshalName(param.MarshalName)
		if e != nil {
			c.JSON(400, gin.H{"message": e.Error()})
			return
		}
		packx := tcpx.NewPackx(marshaller)

		type Result struct {
			MessageID   int32                  `json:"message_id"`
			Header      map[string]interface{} `json:"header"`
			MarshalName string                 `json:"marshal_name"`
			Stream      []byte                 `json:"stream"`
		}
		var results = make([]Result, 0, 10)

		for {
			block, e := packx.FirstBlockOf(reader)
			if e != nil {
				if e == io.EOF {
					break
				}
				fmt.Println(errorx.Wrap(e).Error())
				break
			}
			var result Result
			result.MessageID, e = packx.MessageIDOf(block)
			if e != nil {
				c.JSON(500, gin.H{"message": errorx.Wrap(e)})
				return
			}
			result.MarshalName = packx.Marshaller.MarshalName()
			result.Header, e = packx.HeaderOf(block)
			if e != nil {
				c.JSON(500, gin.H{"message": errorx.Wrap(e)})
				return
			}
			result.Stream, e = packx.BodyBytesOf(block)
			if e != nil {
				c.JSON(500, gin.H{"message": errorx.Wrap(e)})
				return
			}
			results = append(results, result)
		}

		c.JSON(200, gin.H{"message": "success", "blocks": results})
	})

	s := &http.Server{
		Addr:           "7000",
		Handler:        cors.AllowAll().Handler(r),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 21,
	}
	s.ListenAndServe()
}

func Validate() {
	r := gin.Default()
	r.POST("/tcpx/clients/stream/", func(c *gin.Context) {
		type Param struct {
			Stream      []byte `json:"stream"`
			MarshalName string `json:"marshal_name"`
		}
		var param Param
		e := c.Bind(&param)
		if e != nil {
			c.JSON(400, gin.H{"message": e.Error()})
			return
		}
		var user interface{}
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
		switch param.MarshalName {
		case "json":
			user = &JSONUser{}
		case "xml":
			user = &XMLUser{}
		case "toml", "tml":
			user = &TOMLUser{}
		case "yaml", "yml":
			user = &YAMLUser{}
		case "protobuf", "proto":
			user = &model.User{}
		default:
			c.JSON(400, gin.H{"message": "marshal_name only accept ['json', 'xml', 'toml','yaml','protobuf']"})
			return
		}
		message, e := tcpx.UnpackWithMarshallerName(param.Stream, user, param.MarshalName)
		if e != nil {
			c.JSON(400, gin.H{"message": e.Error(), "result": "not ok"})
			return
		}
		c.JSON(200, gin.H{"message": "success", "result": "ok", "ms": message})
	})
	s := &http.Server{
		Addr:           ":7001",
		Handler:        cors.AllowAll().Handler(r),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 21,
	}
	s.ListenAndServe()
}
func Service() {
	srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

	srv.BeforeExit(func() {
		fmt.Println("server stops")
	})
	// If mode is DEBUG, error in framework will log with error spot and time in detail
	// tcpx.SetLogMode(tcpx.DEBUG)

	srv.OnClose = OnClose
	srv.OnConnect = OnConnect

	// Mux routine and OnMessage callback can't meet .
	// When OnMessage is not nil, routes will lose effect.
	// When srv.OnMessage has set, srv.AddHandler() makes no sense, it means user wants to handle raw message stream by self.
	// Besides, if OnMessage is not nil, middlewares of global type(by srv.UseGlobal) and anchor type(by srv.Use, srv.UnUse)
	// will all be executed regardless of an anchor type middleware being unUsed or not.
	// srv.OnMessage = OnMessage

	srv.UseGlobal(MiddlewareGlobal)
	srv.Use("middleware1", Middleware1, "middleware2", Middleware2)
	srv.AddHandler(1, SayHello)

	srv.UnUse("middleware2")
	srv.AddHandler(3, SayGoodBye)

	srv.AddHandler(5, Middleware3, SayName)
	// tcp
	go func() {
		fmt.Println("tcp srv listen on 7171")
		if e := srv.ListenAndServe("tcp", ":7171"); e != nil {
			panic(e)
		}
	}()

	// udp
	go func() {
		fmt.Println("udp srv listen on 7172")
		if e := srv.ListenAndServe("udp", ":7172"); e != nil {
			panic(e)
		}
	}()
	// kcp
	go func() {
		fmt.Println("kcp srv listen on 7173")
		if e := srv.ListenAndServe("kcp", ":7173"); e != nil {
			panic(e)
		}
	}()
}

func OnConnect(c *tcpx.Context) {
	fmt.Println(fmt.Sprintf("connecting from remote host %s network %s", c.ClientIP(), c.Network()))
}
func OnClose(c *tcpx.Context) {
	fmt.Println(fmt.Sprintf("connecting from remote host %s network %s has stoped", c.ClientIP(), c.Network()))
}

var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})

func OnMessage(c *tcpx.Context) {
	type ServiceA struct {
		Username string `json:"username"`
	}
	type ServiceB struct {
		ServiceName string `json:"service_name" toml:"service_name" yaml:"service_name"`
	}

	messageID, e := packx.MessageIDOf(c.Stream)
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}

	switch messageID {
	case 7:
		var serviceA ServiceA
		// block, e := packx.Unpack(c.Stream, &serviceA)
		block, e := c.Bind(&serviceA)
		fmt.Println(block, e)
		c.Reply(8, "success")
	case 9:
		var serviceB ServiceB
		//block, e := packx.Unpack(c.Stream, &serviceB)
		block, e := c.Bind(&serviceB)
		fmt.Println(block, e)
		c.JSON(10, "success")
	}

}
func SayHello(c *tcpx.Context) {
	var messageFromClient string
	var messageInfo tcpx.Message
	messageInfo, e := c.Bind(&messageFromClient)
	if e != nil {
		panic(e)
	}
	fmt.Println("receive messageID:", messageInfo.MessageID)
	fmt.Println("receive header:", messageInfo.Header)
	fmt.Println("receive body:", messageInfo.Body)

	var responseMessageID int32 = 2
	e = c.Reply(responseMessageID, "hello")
	fmt.Println("reply:", "hello")
	if e != nil {
		fmt.Println(e.Error())
	}
}

func SayGoodBye(c *tcpx.Context) {
	var messageFromClient string
	var messageInfo tcpx.Message
	messageInfo, e := c.Bind(&messageFromClient)
	if e != nil {
		panic(e)
	}
	fmt.Println("receive messageID:", messageInfo.MessageID)
	fmt.Println("receive header:", messageInfo.Header)
	fmt.Println("receive body:", messageInfo.Body)

	var responseMessageID int32 = 4
	e = c.Reply(responseMessageID, "bye")
	fmt.Println("reply:", "bye")
	if e != nil {
		fmt.Println(e.Error())
	}
}

func SayName(c *tcpx.Context) {
	var messageFromClient string
	var messageInfo tcpx.Message
	messageInfo, e := c.Bind(&messageFromClient)
	if e != nil {
		panic(e)
	}
	fmt.Println("receive messageID:", messageInfo.MessageID)
	fmt.Println("receive header:", messageInfo.Header)
	fmt.Println("receive body:", messageInfo.Body)

	var responseMessageID int32 = 6
	e = c.Reply(responseMessageID, "my name is tcpx")
	fmt.Println("reply:", "my name is tcpx")
	if e != nil {
		fmt.Println(e.Error())
	}
}

func Middleware1(c *tcpx.Context) {
	fmt.Println("I am middleware 1 exampled by 'srv.Use(\"middleware1\", Middleware1)'")
}

func Middleware2(c *tcpx.Context) {
	fmt.Println("I am middleware 2 exampled by 'srv.Use(\"middleware2\", Middleware2),srv.UnUse(\"middleware2\")'")
}

func Middleware3(c *tcpx.Context) {
	fmt.Println("I am middleware 3 exampled by 'srv.AddHandler(5, Middleware3, SayName)'")
}

func MiddlewareGlobal(c *tcpx.Context) {
	fmt.Println("I am global middleware exampled by 'srv.UseGlobal(MiddlewareGlobal)'")
}
