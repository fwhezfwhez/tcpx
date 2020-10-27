package tcpx

import (
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"

	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const CONTEXT_ONLINE = 1
const CONTEXT_OFFLINE = 2
const ABORT = 2019

// Context has two concurrently safe context:
// PerConnectionContext is used for connection, once the connection is built ,this is connection scope.
// PerRequestContext is used for request, when connection was built, then many requests can be sent between client and server.
// Each request has an independently scoped context , this is PerRequestContext.
// Packx used to save a marshaller helping marshal and unMarshal stream
// Stream is read from net.Conn per request
type Context struct {
	// for tcp conn
	Conn net.Conn
	// context scope lock
	L *sync.RWMutex

	// for udp conn
	PacketConn net.PacketConn
	Addr       net.Addr

	// for kcp conn
	//UDPSession *kcp.UDPSession

	// for k-v pair shared in connection/request scope
	PerConnectionContext *sync.Map
	PerRequestContext    *sync.Map

	// for pack and unpack
	Packx *Packx

	// saves a pool-ref from TcpX instance
	// only when TcpX instance has set builtInPool true, poolRef is not nil
	// - How to use this?
	// `ctx.Online(username)`
	// `ctx.Offline()`
	poolRef *ClientPool

	// signal end, after called `ctx.CloseConn()`, it can broadcast all routine related  to this connection.
	// It will ensure all related goroutine to die.
	recvEnd chan int

	recvAuth chan int

	// 1- online, 2- offline
	// This value will init to 1 by NewContext() and turn 2 by ctx.Close()
	// This value is shared among request context, so it must be pointer int,not value int
	userState *int

	// for raw message
	ConnReader io.Reader
	ConnWriter io.Writer

	// for request scpope,Stream, offset, handlers will be copy when new request comes(same connection)

	Stream []byte
	// used to control middleware abort or next
	// offset == ABORT, abort
	// else next
	offset   int
	handlers []func(*Context)
}

// share some pointer properties With former context, but has independent Stream and handlers
// Not locked. Caller should lock it.
func copyContext(ctx Context) *Context {
	var copyHandlers = make([]func(*Context), len(ctx.handlers))
	for i, _ := range copyHandlers {
		copyHandlers[i] = ctx.handlers[i]
	}

	return &Context{
		Conn:       ctx.Conn,
		L:          ctx.L,
		PacketConn: ctx.PacketConn,
		Addr:       ctx.Addr,
		//UDPSession:           ctx.UDPSession,
		PerConnectionContext: ctx.PerConnectionContext,
		PerRequestContext:    ctx.PerConnectionContext,
		Packx:                ctx.Packx,
		Stream:               ctx.Stream,
		offset:               ctx.offset,
		handlers:             copyHandlers,
		poolRef:              ctx.poolRef,
		recvEnd:              ctx.recvEnd,
		recvAuth:             ctx.recvAuth,
		userState:            ctx.userState,
		ConnReader:           ctx.ConnReader,
		ConnWriter:           ctx.ConnWriter,
	}
}

// No strategy to ensure username repeat or not , if username exists, it will replace the old connection context in the pool.
// Only used when tcpX instance's builtInPool is true,
// otherwise you should design your own client pool(github.com/fwhezfwhez/tcpx/clientPool/client-pool.go), and manage it
// yourself, like:
// ```
//     var myPool = clientPool.NewClientPool()
//     func main() {
//         srv := tcpx.NewTcpX(nil)
//         srv.AddHandler(1, func(c *tcpx.Context){
//             type Login struct{
//                Username string
//             }
//             var userLogin Login
//             c.Bind(&userLogin)
//             myPool.Online(userLogin.Username, c)
//         })
//         srv.AddHandler(2, func(c *tcpx.Context){
//             username, ok := ctx.Username()
//             if !ok {
//                 fmt.Println("anonymous user no need to offline")
//             }
//             myPool.Offline(username)
//         })
//     }
// ```
func (ctx *Context) Online(username string) error {
	if username == "" {
		return errors.New("can't use empty username to online")
	}
	ctx.SetUsername(username)
	if ctx.poolRef == nil {
		return errors.New("ctx.poolRef is nil, did you call 'tcpX.WithBuiltInPool(true)' or 'tcpX.SetPool(pool *tcpx.ClientPool)' yet")
	}
	ctx.poolRef.Online(username, ctx)
	return nil
}

// Only used when tcpX instance's builtInPool is true,
// otherwise you should design your own client pool(github.com/fwhezfwhez/tcpx/clientPool/client-pool.go), and manage it
// yourself, like:
// ```
//     var myPool = clientPool.NewClientPool()
//     func main() {
//         srv := tcpx.NewTcpX(nil)
//         srv.AddHandler(1, func(c *tcpx.Context){
//             type Login struct{
//                Username string
//             }
//             var userLogin Login
//             c.Bind(&userLogin)
//             myPool.Online(userLogin.Username, c)
//         })
//         srv.AddHandler(2, func(c *tcpx.Context){
//             myPool.Offline(userLogin.Username)
//         })
//     }
//```
func (ctx *Context) Offline() error {
	if ctx.poolRef == nil {
		return errors.New("ctx.poolRef is nil, did you call 'tcpX.WithBuiltInPool(true)' or 'tcpX.SetPool(pool *tcpx.ClientPool)' yet")
	}
	username, ok := ctx.Username()
	if !ok {
		return errors.New("offline username  empty, did you call c.Online(username string) yet")
	}
	ctx.poolRef.Offline(username)
	return nil
}

// New a context.
// This is used for new a context for tcp server.
func NewContext(conn net.Conn, marshaller Marshaller) *Context {
	var online = CONTEXT_ONLINE
	return &Context{
		Conn:                 conn,
		PerConnectionContext: &sync.Map{},
		PerRequestContext:    &sync.Map{},

		Packx:  NewPackx(marshaller),
		offset: -1,

		recvEnd:  make(chan int, 1),
		recvAuth: make(chan int, 1),

		L:         &sync.RWMutex{},
		userState: &online,
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
	var online = CONTEXT_ONLINE
	return &Context{
		PacketConn:           conn,
		Addr:                 addr,
		PerConnectionContext: nil,
		PerRequestContext:    &sync.Map{},

		Packx:  NewPackx(marshaller),
		offset: -1,

		recvEnd:  make(chan int, 1),
		recvAuth: make(chan int, 1),

		L:         &sync.RWMutex{},
		userState: &online,
	}
}

// New a context.
// This is used for new a context for kcp server.
//func NewKCPContext(udpSession *kcp.UDPSession, marshaller Marshaller) *Context {
//	var online = CONTEXT_ONLINE
//	return &Context{
//		UDPSession:           udpSession,
//		PerConnectionContext: nil,
//		PerRequestContext:    &sync.Map{},
//
//		Packx:  NewPackx(marshaller),
//		offset: -1,
//
//		recvEnd:  make(chan int, 1),
//		recvAuth: make(chan int, 1),
//
//		L:         &sync.RWMutex{},
//		userState: &online,
//	}
//}

// ConnectionProtocol returns server protocol, tcp, udp, kcp
func (ctx *Context) ConnectionProtocolType() string {
	if ctx.Conn != nil {
		return "tcp"
	}
	if ctx.Addr != nil && ctx.PacketConn != nil {
		return "udp"
	}
	//if ctx.UDPSession != nil {
	//	return "kcp"
	//}
	return "tcp"
}

func (ctx *Context) InitReaderAndWriter() error {
	switch ctx.ConnectionProtocolType() {
	case "tcp":
		ctx.ConnReader = ctx.Conn
		ctx.ConnWriter = ctx.Conn
		//case "kcp":
		//	ctx.ConnReader = ctx.UDPSession
		//	ctx.ConnWriter = ctx.UDPSession

		// udp not support writer and reader
		//case "udp":
		//	ctx.ConnReader = ctx.PacketConn
		//	ctx.ConnWriter = ctx.PacketConn
	default:
		return fmt.Errorf("only accept tcp/kcp but got %s", ctx.ConnectionProtocolType())
	}
	return nil
}

// Close its connection
func (ctx *Context) CloseConn() error {
	defer func() {

		if ctx.recvEnd != nil {
			CloseChanel(func() {
				close(ctx.recvEnd)
			})
		}
		if ctx.poolRef != nil {
			ctx.Offline()
		}

		ctx.L.Lock()
		defer ctx.L.Unlock()
		*(ctx.userState) = CONTEXT_OFFLINE
	}()

	switch ctx.ConnectionProtocolType() {
	case "tcp":
		return ctx.Conn.Close()
	case "udp":

		return ctx.PacketConn.Close()
		//case "kcp":
		//	return ctx.UDPSession.Close()
	}
	return nil
}

// set deadline
func (ctx *Context) SetDeadline(t time.Time) error {
	switch ctx.ConnectionProtocolType() {
	case "tcp":
		return ctx.Conn.SetDeadline(t)
	case "udp":

		return ctx.PacketConn.SetDeadline(t)
		//case "kcp":
		//	return ctx.UDPSession.SetDeadline(t)
	}
	return nil
}

// set read deadline
func (ctx *Context) SetReadDeadline(t time.Time) error {
	switch ctx.ConnectionProtocolType() {
	case "tcp":
		return ctx.Conn.SetReadDeadline(t)
	case "udp":

		return ctx.PacketConn.SetReadDeadline(t)
		//case "kcp":
		//	return ctx.UDPSession.SetReadDeadline(t)
	}
	return nil
}

// set write deadline
func (ctx *Context) SetWriteDeadline(t time.Time) error {
	switch ctx.ConnectionProtocolType() {
	case "tcp":
		return ctx.Conn.SetWriteDeadline(t)
	case "udp":

		return ctx.PacketConn.SetWriteDeadline(t)
		//case "kcp":
		//	return ctx.UDPSession.SetWriteDeadline(t)
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
	if tmp, ok := k.(string); ok {
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

// Context's connection scope saves an unique key to the connection pool
// Before using this, ctx.SetUsername should be call first
func (ctx *Context) Username() (string, bool) {
	usernameI, ok := ctx.GetCtxPerConn("tcpx-username")
	if !ok {
		return "", ok
	}
	return usernameI.(string), ok
}

// When you want to tag an username to the context, use it, or it will be regarded as an anonymous user
func (ctx *Context) SetUsername(username string) {
	ctx.setCtxPerConn("tcpx-username", username)
}

// this has no restriction for key, should be used in local package
func (ctx *Context) setCtxPerConn(k, v interface{}) {
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
func (ctx *Context) ReplyWithMarshaller(marshaller Marshaller, messageID int32, src interface{}, headers ...map[string]interface{}) error {
	return ctx.commonReplyWithMarshaller(marshaller, messageID, src, headers...)
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

func (ctx *Context) commonReplyWithMarshaller(marshaller Marshaller, messageID int32, src interface{}, headers ...map[string]interface{}) error {
	var buf []byte
	var e error
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
		//case "kcp":
		//	if _, e = ctx.UDPSession.Write(buf); e != nil {
		//		return errorx.Wrap(e)
		//	}
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
		//case "kcp":
		//	clientAddr = ctx.UDPSession.RemoteAddr().String()
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

func (ctx *Context) IsOffline() bool {
	ctx.L.RLock()
	defer ctx.L.RUnlock()
	return *(ctx.userState) == CONTEXT_OFFLINE
}

func (ctx *Context) IsOnline() bool {
	if ctx == nil {
		return false
	}

	ctx.L.RLock()
	defer ctx.L.RUnlock()

	return *(ctx.userState) == CONTEXT_ONLINE
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
// It will never be nil, if not exist the channel, it will auto-make.
func (ctx *Context) HeartBeatChan() chan int {
	channel, ok := ctx.GetCtxPerConn("tcpx-heart-beat-channel")
	if !ok {
		channel = make(chan int, 1)
		ctx.setCtxPerConn("tcpx-heart-beat-channel", channel)
		return channel.(chan int)
	} else {
		tmp, ok := channel.(chan int)
		if !ok {
			channel = make(chan int, 1)
			ctx.setCtxPerConn("tcpx-heart-beat-channel", channel)
			return channel.(chan int)
		}
		return tmp
	}
}

// RecvHeartBeat
func (ctx *Context) RecvHeartBeat() {
	ctx.HeartBeatChan() <- 1
}

// Send to another conn index via username.
// Make sure called `srv.WithBuiltInPool(true)`
func (ctx *Context) SendToUsername(username string, messageID int32, src interface{}, headers ...map[string]interface{}) error {
	if ctx.poolRef == nil {
		return errors.New("'ctx.poolRef' is nil, make sure call 'srv.WithBuiltInPool(true)'")
	}
	anotherCtx := ctx.poolRef.GetClientPool(username)
	if anotherCtx == nil || anotherCtx.IsOffline() {
		return errors.New(fmt.Sprintf("username '%s' not found in pool, he/she might get offine", username))
	}
	return ctx.SendToConn(anotherCtx, messageID, src, headers...)
}

// Send to another conn via Context.
// Make sure called `srv.WithBuiltInPool(true)`
func (ctx *Context) SendToConn(anotherCtx *Context, messageID int32, src interface{}, headers ...map[string]interface{}) error {
	return anotherCtx.Reply(messageID, src, headers...)
}

func (ctx *Context) GetPoolRef() *ClientPool {
	ctx.L.RLock()
	defer ctx.L.RUnlock()
	return ctx.poolRef
}

func (ctx *Context) AuthChan() <-chan int {
	return ctx.recvAuth
}

func (ctx *Context) RecvAuthPass() {
	const PASS = 1
	ctx.recvAuth <- PASS
}
func (ctx *Context) RecvAuthDeny() {
	const DENY = -1
	ctx.recvAuth <- DENY
}

func (ctx Context) RouterType() string {
	return ctx.routerType()
}

func (ctx *Context) routerType() string {
	if len(ctx.Stream) == 0 {
		return MESSAGEID
	}

	header, e := HeaderOf(ctx.Stream)
	if e != nil {
		Logger.Println("header decode err: %s", errorx.Wrap(e).Error())
		return MESSAGEID
	}

	if len(header) == 0 {
		return MESSAGEID
	}

	routerTypeI, exist := header[HEADER_ROUTER_KEY]
	if !exist {
		return MESSAGEID
	}

	routerTypeStr, transfer := routerTypeI.(string)
	if !transfer {
		return MESSAGEID
	}

	if routerTypeStr == MESSAGEID {
		return MESSAGEID
	}

	return routerTypeStr

}

func (ctx *Context) JSONURLPattern(urlPattern string, src interface{}) {
	NewURLPatternMessage(urlPattern, src).Pack(JsonMarshaller{})
}
