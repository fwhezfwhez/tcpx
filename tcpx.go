package tcpx

import (
	"errorX"
	"fmt"

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
			Logger.Println(err.Error())
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
			var e error
			for {
				ctx.Stream, e = ctx.Packx.FirstBlockOf(ctx.Conn)
				if e != nil {
					Logger.Println(e)
					break
				}
				ctx.PerRequestContext = &sync.Map{}
				go func(ctx *Context, tcpx *TcpX) {
					if tcpx.OnMessage != nil {
						go tcpx.OnMessage(ctx)
					} else {
						messageID, e := tcpx.Packx.MessageIDOf(ctx.Stream)
						if e != nil {
							Logger.Println(errorx.Wrap(e).Error())
							return
						}
						handler, ok := tcpx.Mux.Handlers[messageID]
						if !ok {
							Logger.Println(fmt.Sprintf("messageID %d handler not found", messageID))
							return
						}
						handler(ctx)
					}
				}(ctx, tcpx)
				continue
			}
		}(ctx, tcpx)
	}
}
