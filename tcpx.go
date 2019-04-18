package tcpx

import (
	"errorX"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// OnMessage and mux are opposite.
// When OnMessage is not nil, users should deal will ctx.Stream themselves.
// When OnMessage is nil, program will handle ctx.Stream via mux routing by messageID
type TcpX struct {
	OnConnect func(ctx *Context)
	OnMessage func(ctx *Context)
	OnClose   func(ctx *Context)
	Mux       *Mux

	Packx *Packx
}

func NewTcpX(marshaller Marshaller) *TcpX {
	return &TcpX{
		Packx: NewPackx(marshaller),
		Mux:   NewMux(),
	}
}

func (tcpx *TcpX) AddHandler(messageID int32, f func(ctx *Context)) {
	tcpx.Mux.AddHandleFunc(messageID, f)
}
func (tcpx *TcpX) ListenAndServe(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		ctx := NewContext(conn, tcpx.Packx.Marshaller)
		if tcpx.OnConnect != nil {
			tcpx.OnConnect(ctx)
		}
		func(ctx *Context, tcpx *TcpX) {
			defer ctx.Conn.Close()
			if tcpx.OnClose != nil {
				defer tcpx.OnClose(ctx)
			}
			// 16 byte info stream
			var info = make([]byte, 16, 16)
			for {
				info = info[:16]
				// content stream
				var content []byte
				fmt.Println("before read from conn, info", info)

				n, e := ctx.Conn.Read(info)
				fmt.Println("after read from conn, info, n", info, n)
				if e != nil {
					fmt.Println(errorx.Wrap(e).Error())
					break
				}
				if n != 16 {
					fmt.Println(errors.New(fmt.Sprintf("read info should be 16 but got %d", n)))
					break
				}
				ctx.PerRequestContext = &sync.Map{}
				contentLength, e := ctx.Packx.LengthOf(info)
				if e != nil {
					fmt.Println(errorx.Wrap(e).Error())
					break
				}
				content = make([]byte, contentLength)

				//var buffer = bytes.NewBuffer(nil)
				//_, e = buffer.ReadFrom(ctx.Conn)
				_, e = ctx.Conn.Read(content)
				if e != nil {
					fmt.Println(errorx.Wrap(e).Error())
					break
				}
				//content = buffer.Bytes()
				ctx.Stream = append(info, content...)

				if tcpx.OnMessage != nil {
					go tcpx.OnMessage(ctx)
				} else {
					messageID, e := tcpx.Packx.MessageIDOf(info)
					if e != nil {
						fmt.Println(errorx.Wrap(e).Error())
						break
					}
					handler, ok := tcpx.Mux.Handlers[messageID]
					if !ok {
						log.Println(fmt.Sprintf("messageID %d handler not found", messageID))
						break
					}
					go handler(ctx)
				}
			}
		}(ctx, tcpx)
	}
}
