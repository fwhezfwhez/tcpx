package tcpx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx/examples/sayHello/client/pb"
	"github.com/xtaci/kcp-go"
	"net"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	tcpCtx := NewContext(&net.TCPConn{}, nil)
	udpCtx := NewUDPContext(&net.UDPConn{}, &net.UDPAddr{}, nil)
	kcpCtx := NewKCPContext(&kcp.UDPSession{}, nil)
	fmt.Println(tcpCtx, udpCtx, kcpCtx)

	if tcpCtx.ConnectionProtocolType() != "tcp" {
		fmt.Println(fmt.Sprintf("tcpCtx want tcp but got %s", tcpCtx.ConnectionProtocolType()))
		t.Fail()
		return
	}
	if udpCtx.ConnectionProtocolType() != "udp" {
		fmt.Println(fmt.Sprintf("udpCtx want udp but got %s", udpCtx.ConnectionProtocolType()))
		t.Fail()
		return
	}
	if kcpCtx.ConnectionProtocolType() != "kcp" {
		fmt.Println(fmt.Sprintf("kcpCtx want kcp but got %s", kcpCtx.ConnectionProtocolType()))
		t.Fail()
		return
	}
}

func TestContext_Bind_JSON(t *testing.T) {
	var payload = struct{ Username string }{"tcpx"}
	var ctx Context
	var e error
	// prepare
	packJson := NewPackx(JsonMarshaller{})
	ctx.Stream, e = packx.Pack(1, payload)
	ctx.Packx = packJson
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	var received struct{ Username string }
	message, e := ctx.Bind(&received)
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	fmt.Println(received)
	fmt.Println(message)
}

func TestContext_Bind_Protobuf(t *testing.T) {
	var payload pb.SayHelloRequest
	payload.Username = "tcpx"
	var ctx Context
	var e error
	// prepare
	packProtobuf := NewPackx(ProtobufMarshaller{})
	ctx.Stream, e = packProtobuf.Pack(1, &payload)
	ctx.Packx = packProtobuf
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	var received pb.SayHelloRequest
	message, e := ctx.Bind(&received)
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	fmt.Println(received)
	fmt.Println(message)
}

func TestContext_CtxPerConnPerRequest(t *testing.T) {
	ctx := NewContext(nil, nil)

	ctx.SetCtxPerConn("username", "tcpx")

	fmt.Println(ctx.GetCtxPerConn("username"))

	ctx.SetCtxPerRequest("password", "123")
	fmt.Println(ctx.GetCtxPerRequest("password"))
}

func TestContext_Reply(t *testing.T) {
	var serverStart = make(chan int, 1)
	var testResult = make(chan error, 1)
	// client
	go func() {
		<-serverStart

		conn, err := net.Dial("tcp", "localhost:7000")
		if err != nil {
			testResult <- errorx.Wrap(err)
			fmt.Println(errorx.Wrap(err).Error())
			return
		}
		go func() {
			for {
				var buf = make([]byte, 512)
				n, e := conn.Read(buf)
				if e != nil {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.Wrap(e).Error())
					return
				}
				var msgFromServer string
				fmt.Println("client receives from server", buf[0:n])
				_, e = PackJSON.Unpack(buf[:n], &msgFromServer)
				if e != nil {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.Wrap(e).Error())
					return
				}
				if msgFromServer != "hello, I'm server" {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.NewFromStringf("want `hello, I'm server` but got `%s`", msgFromServer))
					return
				}
				testResult <- nil
			}
		}()

		buf, e := PackJSON.Pack(1, "hello, I'm client")
		if e != nil {
			testResult <- errorx.Wrap(e)
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		conn.Write(buf)
	}()

	// server
	go func() {
		srv := NewTcpX(JsonMarshaller{})
		srv.OnMessage = func(c *Context) {
			buf, e := c.Packx.Pack(10086, "hello, I'm server")
			if e != nil {
				testResult <- errorx.Wrap(e)
				fmt.Println(errorx.Wrap(e).Error())
				return
			}
			fmt.Println("server sents to client", buf)
			c.Reply(10086, "hello, I'm server")
		}
		go func() {
			time.Sleep(time.Second * 10)
			serverStart <- 1
		}()
		e := srv.ListenAndServe("tcp", ":7000")
		if e != nil {
			testResult <- errorx.Wrap(e)
			fmt.Println(e.Error())
			return
		}
	}()
	e := <-testResult
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
	}
}

func TestContext_JSON(t *testing.T) {
	var serverStart = make(chan int, 1)
	var testResult = make(chan error, 1)
	// client
	go func() {
		<-serverStart

		conn, err := net.Dial("tcp", "localhost:7001")
		if err != nil {
			testResult <- errorx.Wrap(err)
			fmt.Println(errorx.Wrap(err).Error())
			return
		}
		go func() {
			for {
				var buf = make([]byte, 512)
				n, e := conn.Read(buf)
				if e != nil {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.Wrap(e).Error())
					return
				}
				var msgFromServer string
				fmt.Println("client receives from server", buf[0:n])
				_, e = PackJSON.Unpack(buf[:n], &msgFromServer)
				if e != nil {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.Wrap(e).Error())
					return
				}
				if msgFromServer != "hello, I'm server" {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.NewFromStringf("want `hello, I'm server` but got `%s`", msgFromServer))
					return
				}
				testResult <- nil
			}
		}()

		buf, e := PackJSON.Pack(1, "hello, I'm client")
		if e != nil {
			testResult <- errorx.Wrap(e)
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		conn.Write(buf)
	}()

	// server
	go func() {
		srv := NewTcpX(JsonMarshaller{})
		srv.OnMessage = func(c *Context) {
			buf, e := c.Packx.Pack(10086, "hello, I'm server")
			if e != nil {
				testResult <- errorx.Wrap(e)
				fmt.Println(errorx.Wrap(e).Error())
				return
			}
			fmt.Println("server sents to client", buf)
			c.Reply(10086, "hello, I'm server")
		}
		go func() {
			time.Sleep(time.Second * 10)
			serverStart <- 1
		}()
		e := srv.ListenAndServe("tcp", ":7001")
		if e != nil {
			testResult <- errorx.Wrap(e)
			fmt.Println(e.Error())
			return
		}
	}()
	e := <-testResult
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
	}
}

func TestContext_ProtoBuf(t *testing.T) {
	var serverStart = make(chan int, 1)
	var testResult = make(chan error, 1)
	// client
	go func() {
		<-serverStart

		conn, err := net.Dial("tcp", "localhost:7002")
		if err != nil {
			testResult <- errorx.Wrap(err)
			fmt.Println(errorx.Wrap(err).Error())
			return
		}
		go func() {
			for {
				var buf = make([]byte, 512)
				n, e := conn.Read(buf)
				if e != nil {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.Wrap(e).Error())
					return
				}
				var msgFromServer pb.SayHelloReponse
				//fmt.Println("client receives from server" ,buf[0:n])
				_, e = PackProtobuf.Unpack(buf[:n], &msgFromServer)
				if e != nil {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.Wrap(e).Error())
					return
				}
				if msgFromServer.Message != "hello, I'm server" {
					testResult <- errorx.Wrap(e)
					fmt.Println(errorx.NewFromStringf("want `hello, I'm server` but got `%s`", msgFromServer.Message))
					return
				}
				testResult <- nil
			}
		}()

		buf, e := PackJSON.Pack(1, "hello, I'm client")
		if e != nil {
			testResult <- errorx.Wrap(e)
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		conn.Write(buf)
	}()

	// server
	go func() {
		srv := NewTcpX(JsonMarshaller{})
		srv.OnMessage = func(c *Context) {
			var response pb.SayHelloReponse
			response.Message = "hello, I'm server"
			fmt.Println(c.ClientIP())
			c.ProtoBuf(10086, &response)
		}
		go func() {
			time.Sleep(time.Second * 10)
			serverStart <- 1
		}()
		e := srv.ListenAndServe("tcp", ":7002")
		if e != nil {
			testResult <- errorx.Wrap(e)
			fmt.Println(e.Error())
			return
		}
	}()
	e := <-testResult
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
	}
}

func TestContext_Network(t *testing.T) {
	ctx := NewTCPContext(&net.TCPConn{}, nil)
	if ctx.Network() != "tcp" {
		fmt.Println(fmt.Sprintf("ctx want tcp but got %s", ctx.Network()))
		t.Fail()
		return
	}
	ctx = NewUDPContext(&net.UDPConn{}, &net.UDPAddr{}, nil)
	if ctx.Network() != "udp" {
		fmt.Println(fmt.Sprintf("ctx want udp but got %s", ctx.Network()))
		t.Fail()
		return
	}
	ctx = NewKCPContext(&kcp.UDPSession{}, nil)
	if ctx.Network() != "kcp" {
		fmt.Println(fmt.Sprintf("ctx want kcp but got %s", ctx.Network()))
		t.Fail()
		return
	}
}

func TestContext_Abort_Next_Reset_RestOffset_IsAbort(t *testing.T) {
	ctx := NewTCPContext(nil, nil)
	ctx.Abort()
	if ctx.offset != ABORT {
		fmt.Println(fmt.Sprintf("offset want %d but got %d", ABORT, ctx.offset))
		t.Fail()
		return
	}
	if ctx.isAbort() != true {
		fmt.Println(fmt.Sprintf("ctx.isAbort want %v but got %v", true, ctx.isAbort()))
		t.Fail()
		return
	}
	ctx.Reset()
	if ctx.offset != -1 {
		fmt.Println(fmt.Sprintf("offset want %d but got %d", -1, ctx.offset))
		t.Fail()
		return
	}
}
