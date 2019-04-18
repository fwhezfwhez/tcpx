package tcpx

import (
	"errorX"
	"net"
	"sync"
)

// Context has two concurrently safe context:
// PerConnectionContext is used for connection, once the connection is built ,this is connection scope.
// PerRequestContext is used for request, when connection was built, then many requests can be sent between client and server.
// Each request has an independently scoped context , this is PerRequestContext.
// Packx used to save a marshaller helping marshal and unMarshal stream
// Stream is read from net.Conn per request
type Context struct {
	Conn                 net.Conn
	PerConnectionContext *sync.Map
	PerRequestContext    *sync.Map

	Packx  *Packx
	Stream []byte
}

func NewContext(conn net.Conn, marshaller Marshaller) *Context {
	return &Context{
		Conn:                 conn,
		PerConnectionContext: &sync.Map{},
		PerRequestContext:    &sync.Map{},

		Packx: NewPackx(marshaller),
	}
}
func (ctx *Context) Bind(dest interface{}) (Message, error) {
	return ctx.Packx.Unpack(ctx.Stream, dest)
}

func (ctx *Context) SetCtxPerConn(k, v interface{}) {
	ctx.PerConnectionContext.Store(k, v)
}

func (ctx *Context) GetCtxPerConn(k interface{}) (interface{}, bool) {
	return ctx.PerConnectionContext.Load(k)
}

func (ctx *Context) SetCtxPerRequest(k, v interface{}) {
	ctx.PerRequestContext.Store(k, v)
}

func (ctx *Context) GetCtxPerRequest(k interface{}) (interface{}, bool) {
	return ctx.PerRequestContext.Load(k)
}
func (ctx *Context) Reply(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	var buf []byte
	var e error
	buf, e = ctx.Packx.Pack(messageID, src, headers ...)
	if _, e = ctx.Conn.Write(buf); e != nil {
		return errorx.Wrap(e)
	}

	return nil
}

func (ctx *Context) JSON(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	var buf []byte
	var e error
	if ctx.Packx.Marshaller.MarshalName() != "json" {
		buf, e = NewPackx(JsonMarshaller{}).Pack(messageID, src, headers...)
		if e != nil {
			return errorx.Wrap(e)
		}
		_, e = ctx.Conn.Write(buf)
		if e != nil {
			return errorx.Wrap(e)
		}
	}
	buf, e = ctx.Packx.Pack(messageID, src, headers ...)
	if _, e = ctx.Conn.Write(buf); e != nil {
		return errorx.Wrap(e)
	}

	return nil
}
