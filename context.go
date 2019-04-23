package tcpx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"net"
	"strings"
	"sync"
)

const (
	ABORT = 2019
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

	// used to control middleware abort or next
	// offset == ABORT, abort
	// else next
	offset   int
	handlers []func(*Context)
}

func NewContext(conn net.Conn, marshaller Marshaller) *Context {
	return &Context{
		Conn:                 conn,
		PerConnectionContext: &sync.Map{},
		PerRequestContext:    &sync.Map{},

		Packx:  NewPackx(marshaller),
		offset: -1,
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

// Reply to client using ctx's well-set Packx.Marshaller.
func (ctx *Context) Reply(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	var buf []byte
	var e error
	buf, e = ctx.Packx.Pack(messageID, src, headers ...)
	if _, e = ctx.Conn.Write(buf); e != nil {
		return errorx.Wrap(e)
	}

	return nil
}

// Reply to client using json marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'json' or not , message block will marshal its header and body by json marshaller.
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

// not finished
func (ctx *Context) XML(messageID int32, src interface{}, headers ...map[string]interface{}) error {
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

// not finished
func (ctx *Context) TOML(messageID int32, src interface{}, headers ...map[string]interface{}) error {
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

// not finished
func (ctx *Context) YAML(messageID int32, src interface{}, headers ...map[string]interface{}) error {
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

// not finished
func (ctx *Context) ProtoBuf(messageID int32, src interface{}, headers ...map[string]interface{}) error {
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

// client ip
func (ctx Context) ClientIP() string {
	arr := strings.Split(ctx.Conn.RemoteAddr().String(), ":")
	// ipv4
	if len(arr) == 2 {
		return arr[0]
	}
	// [::1] 本机
	if strings.Contains(ctx.Conn.RemoteAddr().String(), "[") && strings.Contains(ctx.Conn.RemoteAddr().String(), "]") {
		return "127.0.0.1"
	}
	// ivp6
	return strings.Join(arr[:len(arr)-1], ":")
}

// stop middleware chain
func (ctx *Context) Abort() {
	ctx.offset = ABORT
}

// Since middlewares are divided into 3 kinds: global, messageIDSelfRelated, anchorType,
// offset can't be used straightly to control middlewares like  middlewares[offset]().
// Thus, c.Next() means actually do nothing.
func (ctx *Context) Next() {
	ctx.offset ++
	fmt.Println(ctx.offset)
	s := len(ctx.handlers)
	fmt.Println(s)
	for ; ctx.offset < s; ctx.offset++ {
		if !ctx.isAbort() {
			ctx.handlers[ctx.offset](ctx)
		} else{
			return
		}
	}
}
func (ctx *Context) ResetOffset() {
	ctx.offset = -1
}

func (ctx *Context) Reset() {
	ctx.PerRequestContext = &sync.Map{}
	ctx.offset = -1
	if ctx.handlers == nil {
		ctx.handlers = make([]func(*Context), 0, 10)
		return
	}
	ctx.handlers = ctx.handlers[:0]
}
func (ctx *Context) isAbort() bool{
	if ctx.offset >= ABORT {
		return true
	}
	return false
}
