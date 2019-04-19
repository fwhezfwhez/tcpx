package tcpx

import (
	"errorX"
	"fmt"
	"io"
	"reflect"

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
	Packx     *Packx
}

func NewTcpX(marshaller Marshaller) *TcpX {
	return &TcpX{
		Packx: NewPackx(marshaller),
		Mux:   NewMux(),
	}
}

func (tcpx *TcpX) Use(mids ... interface{}) {
	if tcpx.Mux == nil {
		tcpx.Mux = NewMux()
	}

	if len(mids)%2 != 0 {
		panic(errorx.NewFromStringf("tcpx.Use(mids ...),'mids' should show in pairs,but got length(mids) %d", len(mids)))
	}
	var middlewareKey string
	var ok bool
	var middleware func(c *Context)

	var middlewareAnchor MiddlewareAnchor
	for i := 0; i < len(mids)-1; i++ {
		for j := i + 1; j < len(mids); j++ {
			middlewareKey, ok = mids[i].(string)
			if !ok {
				panic(errorx.NewFromStringf("tcpx.Use(mids ...), 'mids' index '%d' should be string key type but got %v", i, mids[i]))
			}
			middleware, ok = mids[j].(func(c *Context))
			if !ok {
				panic(errorx.NewFromStringf("tcpx.Use(mids ...), 'mids' index '%d' should be func(c *tcpx.Context) type but got %s", j, reflect.TypeOf(mids[j]).Kind().String()))
			}
			middlewareAnchor.Middleware = middleware
			middlewareAnchor.MiddlewareKey = middlewareKey
			middlewareAnchor.AnchorIndex = tcpx.Mux.CurrentAnchorIndex()
			middlewareAnchor.ExpireAnchorIndex = NOT_EXPIRE

			tcpx.Mux.AddMiddlewareAnchor(middlewareAnchor)
		}
	}
}
func (tcpx *TcpX) UnUse(middlewareKeys ...string) {
	var middlewareAnchor MiddlewareAnchor
	var ok bool
	for _, k := range middlewareKeys {
		if middlewareAnchor, ok = tcpx.Mux.MiddlewareAnchorMap[k]; !ok {
			panic(errorx.NewFromStringf("middlewareKey '%s' not found in mux.MiddlewareAnchorMap", k))
		}
		middlewareAnchor.ExpireAnchorIndex = tcpx.Mux.CurrentAnchorIndex()
		tcpx.Mux.ReplaceMiddlewareAnchor(middlewareAnchor)
	}
}

func (tcpx *TcpX) AddHandler(messageID int32, handlers ... func(ctx *Context)) {
	if len(handlers) <=0 {
		panic(errorx.NewFromStringf("handlers should more than 1 but got %d", len(handlers)))
	}
	if len(handlers) >1 {
		tcpx.Mux.AddMessageIDSelfMiddleware(messageID, handlers[:len(handlers)-1]...)
	}

	f := handlers[len(handlers)-1]
	if tcpx.Mux == nil {
		tcpx.Mux = NewMux()
	}
	tcpx.Mux.AddHandleFunc(messageID, f)
	var messageIDAnchor MessageIDAnchor
	messageIDAnchor.MessageID = messageID
	messageIDAnchor.AnchorIndex = tcpx.Mux.CurrentAnchorIndex()
	tcpx.Mux.AddMessageIDAnchor(messageIDAnchor)
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
		go func(ctx *Context, tcpx *TcpX) {
			defer ctx.Conn.Close()
			if tcpx.OnClose != nil {
				defer tcpx.OnClose(ctx)
			}
			var e error
			for {
				ctx.Stream, e = ctx.Packx.FirstBlockOf(ctx.Conn)
				if e != nil {
					if e == io.EOF {
						break
					}
					Logger.Println(e)
					break
				}
				ctx.PerRequestContext = &sync.Map{}

				go func(ctx *Context, tcpx *TcpX) {
					if tcpx.OnMessage != nil {
						tcpx.Mux.execAllMiddlewares(ctx)
						tcpx.OnMessage(ctx)
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

						tcpx.Mux.execMessageIDMiddlewares(ctx, messageID)
						handler(ctx)
					}
				}(ctx, tcpx)
				continue
			}
		}(ctx, tcpx)
	}
}
