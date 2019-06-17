package tcpx

import (
	"github.com/fwhezfwhez/errorx"
	"github.com/xtaci/kcp-go"
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
	// for tcp conn
	Conn net.Conn

	// for udp conn
	PacketConn net.PacketConn
	Addr       net.Addr

	// for kcp conn
	UDPSession *kcp.UDPSession

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

// New a context.
// This is used for new a context for tcp server.
func NewContext(conn net.Conn, marshaller Marshaller) *Context {
	return &Context{
		Conn:                 conn,
		PerConnectionContext: &sync.Map{},
		PerRequestContext:    &sync.Map{},

		Packx:  NewPackx(marshaller),
		offset: -1,
	}
}

// New a context.
// This is used for new a context for tcp server.
func NewTCPContext(conn net.Conn, marshaller Marshaller) *Context {
	return NewContext(conn, marshaller)
}

// New a context.
// This is used for new a context for udp server.
func NewUDPContext(conn net.PacketConn, addr net.Addr, marshaller Marshaller) *Context {
	return &Context{
		PacketConn:           conn,
		Addr:                 addr,
		PerConnectionContext: nil,
		PerRequestContext:    &sync.Map{},

		Packx:  NewPackx(marshaller),
		offset: -1,
	}
}

// New a context.
// This is used for new a context for kcp server.
func NewKCPContext(udpSession *kcp.UDPSession, marshaller Marshaller) *Context {
	return &Context{
		UDPSession:           udpSession,
		PerConnectionContext: nil,
		PerRequestContext:    &sync.Map{},

		Packx:  NewPackx(marshaller),
		offset: -1,
	}
}

// ConnectionProtocol returns server protocol, tcp, udp, kcp
func (ctx *Context) ConnectionProtocolType() string {
	if ctx.Conn != nil {
		return "tcp"
	}
	if ctx.Addr != nil && ctx.PacketConn != nil {
		return "udp"
	}
	if ctx.UDPSession != nil {
		return "kcp"
	}
	return "tcp"
}

// Close its connection
func (ctx *Context) CloseConn() error{
	switch ctx.ConnectionProtocolType() {
	case "tcp":
		return ctx.Conn.Close()
	case "udp":
		return ctx.PacketConn.Close()
	case "kcp":
		return ctx.UDPSession.Close()
	}
	return nil
}
func (ctx *Context) Bind(dest interface{}) (Message, error) {
	return ctx.Packx.Unpack(ctx.Stream, dest)
}

// When context serves for tcp, set context k-v pair of PerConnectionContext.
// When context serves for udp, set context k-v pair of PerRequestContext
// Key should not start with 'tcpx-', or it will panic.
func (ctx *Context) SetCtxPerConn(k, v interface{}) {
	if tmp,ok :=k.(string);ok{
		if strings.HasPrefix(tmp, "tcpx-") {
			panic("keys starting with 'tcpx-' are not allowed setting, they're used officially inside")
		}
	}

	if ctx.ConnectionProtocolType() == "udp" {
		ctx.SetCtxPerRequest(k, v)
		return
	}
	ctx.PerConnectionContext.Store(k, v)
}

// When context serves for tcp, get context k-v pair of PerConnectionContext.
// When context serves for udp, get context k-v pair of PerRequestContext.
func (ctx *Context) GetCtxPerConn(k interface{}) (interface{}, bool) {
	if ctx.ConnectionProtocolType() == "udp" {
		return ctx.GetCtxPerRequest(k)
	}
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
	if e != nil {
		return errorx.Wrap(e)
	}
	return ctx.replyBuf(buf)
}

// Reply to client using json marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'json' or not , message block will marshal its header and body by json marshaller.
func (ctx *Context) JSON(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	return ctx.commonReply("json", messageID, src, headers...)
}

// Reply to client using xml marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'xml' or not , message block will marshal its header and body by xml marshaller.
func (ctx *Context) XML(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	return ctx.commonReply("xml", messageID, src, headers...)
}

// Reply to client using toml marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'toml' or not , message block will marshal its header and body by toml marshaller.
func (ctx *Context) TOML(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	return ctx.commonReply("toml", messageID, src, headers...)
}

// Reply to client using yaml marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'yaml' or not , message block will marshal its header and body by yaml marshaller.
func (ctx *Context) YAML(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	return ctx.commonReply("yaml", messageID, src, headers...)
}

// Reply to client using protobuf marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'protobuf' or not , message block will marshal its header and body by protobuf marshaller.
func (ctx *Context) ProtoBuf(messageID int32, src interface{}, headers ...map[string]interface{}) error {
	return ctx.commonReply("protobuf", messageID, src, headers...)
}

func (ctx *Context) commonReply(marshalName string, messageID int32, src interface{}, headers ...map[string]interface{}) error {
	var buf []byte
	var e error
	var marshaller Marshaller
	if ctx.Packx.Marshaller.MarshalName() != marshalName {
		marshaller, e = GetMarshallerByMarshalName(marshalName)
		if e != nil {
			return errorx.Wrap(e)
		}
		buf, e = NewPackx(marshaller).Pack(messageID, src, headers...)
		if e != nil {
			return errorx.Wrap(e)
		}
		e = ctx.replyBuf(buf)
		if e != nil {
			return errorx.Wrap(e)
		}
		return nil
	}
	buf, e = ctx.Packx.Pack(messageID, src, headers ...)
	if e != nil {
		return errorx.Wrap(e)
	}
	return ctx.replyBuf(buf)
}

// Divide to udp and tcp replying accesses.
func (ctx *Context) replyBuf(buf []byte) (e error) {
	switch ctx.ConnectionProtocolType() {
	case "tcp":
		if _, e = ctx.Conn.Write(buf); e != nil {
			return errorx.Wrap(e)
		}
	case "udp":
		if _, e = ctx.PacketConn.WriteTo(buf, ctx.Addr); e != nil {
			return errorx.Wrap(e)
		}
	case "kcp":
		if _, e = ctx.UDPSession.Write(buf); e != nil {
			return errorx.Wrap(e)
		}
	}
	return nil
}

func (ctx Context) Network() string {
	return ctx.ConnectionProtocolType()
}

// client ip
func (ctx Context) ClientIP() string {
	var clientAddr string
	switch ctx.ConnectionProtocolType() {
	case "tcp":
		clientAddr = ctx.Conn.RemoteAddr().String()
	case "udp":
		clientAddr = ctx.Addr.String()
	case "kcp":
		clientAddr = ctx.UDPSession.RemoteAddr().String()
	}
	arr := strings.Split(clientAddr, ":")
	// ipv4
	if len(arr) == 2 {
		return arr[0]
	}
	// [::1] localhost
	if strings.Contains(clientAddr, "[") && strings.Contains(clientAddr, "]") {
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
	s := len(ctx.handlers)
	for ; ctx.offset < s; ctx.offset++ {
		if !ctx.isAbort() {
			ctx.handlers[ctx.offset](ctx)
		} else {
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
func (ctx *Context) isAbort() bool {
	if ctx.offset >= ABORT {
		return true
	}
	return false
}

// BindWithMarshaller will specific marshaller.
// in contract, c.Bind() will use its inner packx object marshaller
func (ctx *Context) BindWithMarshaller(dest interface{}, marshaller Marshaller) (Message, error) {
	return NewPackx(marshaller).Unpack(ctx.Stream, dest)
}

// ctx.Stream is well marshaled by pack tool.
// ctx.RawStream is help to access raw stream.
func (ctx *Context) RawStream() ([]byte, error) {
	return ctx.Packx.BodyBytesOf(ctx.Stream)
}

// HeartBeatChan returns a prepared chan int to save heart-beat signal.
func (ctx *Context) HeartBeatChan() chan int {
	channel, ok :=ctx.GetCtxPerConn("tcpx-heart-beat-channel")
	if !ok {
		channel = make(chan int ,1)
		ctx.SetCtxPerConn("tcpx-heart-beat-channel", channel)
		return channel.(chan int)
	} else{
        tmp,ok := channel.(chan int)
        if !ok {
			channel = make(chan int ,1)
			ctx.SetCtxPerConn("tcpx-heart-beat-channel", channel)
			return channel.(chan int)
		}
        return tmp
	}
}