// Package tcpx provides udp,tcp,kcp three kinds of protocol.
package tcpx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/xtaci/kcp-go"
	"io"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"net"
)

const (
	DEFAULT_HEARTBEAT_MESSAGEID = - 1392
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

	// heartbeat setting
	HeartBeatOn        bool          // whether start a goroutine to spy on each connection
	HeatBeatInterval   time.Duration // heartbeat should receive in the interval
	HeartBeatMessageID int32         // which messageID to listen to heartbeat
	ThroughMiddleware  bool          // whether heartbeat go through middleware
}

// new an tcpx srv instance
func NewTcpX(marshaller Marshaller) *TcpX {
	return &TcpX{
		Packx: NewPackx(marshaller),
		Mux:   NewMux(),
	}
}

// Set built in heart beat on
// Default heartbeat handler will be added by messageID tcpx.DEFAULT_HEARTBEAT_MESSAGEID(-1392),
// and default heartbeat handler will not execute all kinds of middleware.
//
// ...
// srv := tcpx.NewTcpX(nil)
// srv.HeartBeatMode(true, 10 * time.Second)
// ...
//
// * If you want specific official heartbeat handler messageID and make it execute middleware:
// srv.HeartBeatModeDetail(true, 10 * time.Second, true, 1)
//
// * If you want to rewrite heartbeat handler:
// srv.RewriteHeartBeatHandler(func(c *tcpx.Context){})
func (tcpx *TcpX) HeartBeatMode(on bool, duration time.Duration) {
	tcpx.HeartBeatOn = on
	tcpx.HeatBeatInterval = duration
	tcpx.ThroughMiddleware = false
	tcpx.HeartBeatMessageID = DEFAULT_HEARTBEAT_MESSAGEID

	if on {
		tcpx.AddHandler(DEFAULT_HEARTBEAT_MESSAGEID, func(c *Context) {
			Logger.Println("recv heartbeat:", c.Stream)

			c.RecvHeartBeat()
		})
	}
}

// specific args for heartbeat
func (tcpx *TcpX) HeartBeatModeDetail(on bool, duration time.Duration, throughMiddleware bool, messageID int32) {
	tcpx.HeartBeatOn = on
	tcpx.HeatBeatInterval = duration
	tcpx.ThroughMiddleware = throughMiddleware
	tcpx.HeartBeatMessageID = messageID

	if on {
		tcpx.AddHandler(messageID, func(c *Context) {
			Logger.Println("recv heartbeat:", c.Stream)
			c.RecvHeartBeat()
		})
	}
}

// Rewrite heartbeat handler
// It will inherit properties of the older heartbeat handler:
//   * heartbeatInterval
//   * throughMiddleware
func (tcpx *TcpX) RewriteHeartBeatHandler(messageID int32, f func(c *Context)) {
	tcpx.removeHandler(tcpx.HeartBeatMessageID)
	tcpx.HeartBeatMessageID = messageID
	tcpx.AddHandler(messageID, f)
}

// remove a handler by messageID.
// this method is used for rewrite heartbeat handler
func (tcpx *TcpX) removeHandler(messageID int32) {
	delete(tcpx.Mux.Handlers, messageID)
	delete(tcpx.Mux.MessageIDAnchorMap, messageID)
}

// Middleware typed 'AnchorTypedMiddleware'.
// Add middlewares ruled by (string , func(c *Context),string , func(c *Context),string , func(c *Context)...).
// Middlewares will be added with an indexed key, which is used to unUse this middleware.
// Each middleware added will be well set an anchor index, when UnUse this middleware, its expire_anchor_index will be well set too.
func (tcpx *TcpX) Use(mids ...interface{}) {
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
	for i := 0; i < len(mids)-1; i += 2 {
		j := i + 1
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

// UnUse an middleware.
// a unused middleware will expired among handlers added after it.For example:
//
// 	srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
//  srv.Use("middleware1", Middleware1, "middleware2", Middleware2)
//	srv.AddHandler(1, SayHello)
//	srv.UnUse("middleware2")
//	srv.AddHandler(3, SayGoodBye)
//
// middleware1 and middleware2 will both work to handler 'SayHello'.
// middleware1 will work to handler 'SayGoodBye' but middleware2 will not work to handler 'SayGoodBye'
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

// Middleware typed 'GlobalTypedMiddleware'.
// GlobalMiddleware will work to all handlers.
func (tcpx *TcpX) UseGlobal(mids ...func(c *Context)) {
	if tcpx.Mux == nil {
		tcpx.Mux = NewMux()
	}
	tcpx.Mux.AddGlobalMiddleware(mids ...)
}

// Middleware typed 'SelfRelatedTypedMiddleware'.
// Add handlers routing by messageID
func (tcpx *TcpX) AddHandler(messageID int32, handlers ...func(ctx *Context)) {
	if len(handlers) <= 0 {
		panic(errorx.NewFromStringf("handlers should more than 1 but got %d", len(handlers)))
	}
	if len(handlers) > 1 {
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

// Start to listen.
// Serve can decode stream generated by packx.
// Support tcp and udp
func (tcpx *TcpX) ListenAndServe(network, addr string) error {
	if In(network, []string{"tcp", "tcp4", "tcp6", "unix", "unixpacket"}) {
		return tcpx.ListenAndServeTCP(network, addr)
	}
	if In(network, []string{"udp", "udp4", "udp6", "unixgram", "ip%"}) {
		return tcpx.ListenAndServeUDP(network, addr)
	}
	if In(network, []string{"kcp"}) {
		return tcpx.ListenAndServeKCP(network, addr)
	}
	return errorx.NewFromStringf("'network' doesn't support '%s'", network)
}

// tcp
func (tcpx *TcpX) ListenAndServeTCP(network, addr string) error {
	defer func() {
		if e := recover(); e != nil {
			Logger.Println(fmt.Sprintf("recover from panic %v", e))
			return
		}
	}()
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
		go heartBeatWatch(ctx, tcpx)

		go func(ctx *Context, tcpx *TcpX) {
			defer func() {
				if e := recover(); e != nil {
					Logger.Println(fmt.Sprintf("recover from panic %v", e))
				}
			}()
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

				// Since ctx.handlers and ctx.offset will change per request, cannot take this function as a new routine,
				// or ctx.offset and ctx.handler will get dirty
				//func(ctx *Context, tcpx *TcpX) {
				//	if tcpx.OnMessage != nil {
				//		// tcpx.Mux.execAllMiddlewares(ctx)
				//		//tcpx.OnMessage(ctx)
				//		if ctx.handlers == nil {
				//			ctx.handlers = make([]func(c *Context), 0, 10)
				//		}
				//		ctx.handlers = append(ctx.handlers, tcpx.Mux.GlobalMiddlewares...)
				//		//for _, v := range tcpx.Mux.MiddlewareAnchorMap {
				//		//	ctx.handlers = append(ctx.handlers, v.Middleware)
				//		//}
				//		for _, v := range tcpx.Mux.MiddlewareAnchors {
				//				ctx.handlers = append(ctx.handlers, v.Middleware)
				//		}
				//		ctx.handlers = append(ctx.handlers, tcpx.OnMessage)
				//		if len(ctx.handlers) > 0 {
				//			ctx.Next()
				//		}
				//		ctx.Reset()
				//	} else {
				//		messageID, e := tcpx.Packx.MessageIDOf(ctx.Stream)
				//		if e != nil {
				//			Logger.Println(errorx.Wrap(e).Error())
				//			return
				//		}
				//		handler, ok := tcpx.Mux.Handlers[messageID]
				//		if !ok {
				//			Logger.Println(fmt.Sprintf("messageID %d handler not found", messageID))
				//			return
				//		}
				//
				//		//handler(ctx)
				//
				//		if ctx.handlers == nil {
				//			ctx.handlers = make([]func(c *Context), 0, 10)
				//		}
				//
				//		// global middleware
				//		ctx.handlers = append(ctx.handlers, tcpx.Mux.GlobalMiddlewares...)
				//		// anchor middleware
				//		messageIDAnchorIndex := tcpx.Mux.AnchorIndexOfMessageID(messageID)
				//		for _, v := range tcpx.Mux.MiddlewareAnchors {
				//			if messageIDAnchorIndex > v.AnchorIndex && messageIDAnchorIndex <= v.ExpireAnchorIndex {
				//				ctx.handlers = append(ctx.handlers, v.Middleware)
				//			}
				//		}
				//		// self-related middleware
				//		ctx.handlers = append(ctx.handlers, tcpx.Mux.MessageIDSelfMiddleware[messageID]...)
				//		// handler
				//		ctx.handlers = append(ctx.handlers, handler)
				//
				//		if len(ctx.handlers) > 0 {
				//			ctx.Next()
				//		}
				//		ctx.Reset()
				//	}
				//}(ctx, tcpx)
				handleMiddleware(ctx, tcpx)
				continue
			}
		}(ctx, tcpx)
	}
}

// udp
// maxBufferSize can set buffer length, if receive a message longer than it ,
func (tcpx *TcpX) ListenAndServeUDP(network, addr string, maxBufferSize ...int) error {
	if len(maxBufferSize) > 1 {
		panic(errorx.NewFromStringf("'tcpx.ListenAndServeUDP''s maxBufferSize should has length less by 1 but got %d", len(maxBufferSize)))
	}

	conn, err := net.ListenPacket(network, addr)
	if err != nil {
		panic(err)
	}

	// listen to incoming udp packets
	go func(conn net.PacketConn, tcpx *TcpX) {
		defer func() {
			if e := recover(); e != nil {
				Logger.Println(fmt.Sprintf("recover from panic %v", e))
			}
		}()
		var buffer []byte
		var addr net.Addr
		var e error
		for {
			// read from udp conn
			buffer, addr, e = ReadAllUDP(conn, maxBufferSize...)
			// global
			if e != nil {
				if e == io.EOF {
					break
				}
				Logger.Println(e.Error())
				continue
				//conn.Close()
				//conn, err = net.ListenPacket(network, addr)
				//if err != nil {
				//	panic(err)
				//}
			}
			ctx := NewUDPContext(conn, addr, tcpx.Packx.Marshaller)

			go heartBeatWatch(ctx, tcpx)

			ctx.Stream, e = tcpx.Packx.FirstBlockOfBytes(buffer)
			if e != nil {
				Logger.Println(e.Error())
				break
			}
			// This function are shared among udp ListenAndServe,tcp ListenAndServe and kcp ListenAndServe.
			// But there are some important differences.
			// tcp's context is per-connection scope, some middleware offset and temporary handlers are saved in
			// this context,which means, this function can't work in parallel goroutines.But udp's context is
			// per-request scope, middleware's args are request-apart, it can work in parallel goroutines because
			// different request has different context instance.It's concurrently safe.
			// Thus we can use it like : `go func(ctx *Context, tcpx *TcpX){...}(ctx, tcpx)`
			//go func(ctx *Context, tcpx *TcpX) {
			//	if tcpx.OnMessage != nil {
			//		// tcpx.Mux.execAllMiddlewares(ctx)
			//		//tcpx.OnMessage(ctx)
			//		if ctx.handlers == nil {
			//			ctx.handlers = make([]func(c *Context), 0, 10)
			//		}
			//		ctx.handlers = append(ctx.handlers, tcpx.Mux.GlobalMiddlewares...)
			//		for _, v := range tcpx.Mux.MiddlewareAnchors {
			//			ctx.handlers = append(ctx.handlers, v.Middleware)
			//		}
			//		ctx.handlers = append(ctx.handlers, tcpx.OnMessage)
			//		if len(ctx.handlers) > 0 {
			//			ctx.Next()
			//		}
			//		ctx.Reset()
			//	} else {
			//		messageID, e := tcpx.Packx.MessageIDOf(ctx.Stream)
			//		if e != nil {
			//			Logger.Println(errorx.Wrap(e).Error())
			//			return
			//		}
			//		handler, ok := tcpx.Mux.Handlers[messageID]
			//		if !ok {
			//			Logger.Println(fmt.Sprintf("messageID %d handler not found", messageID))
			//			return
			//		}
			//
			//		//handler(ctx)
			//
			//		if ctx.handlers == nil {
			//			ctx.handlers = make([]func(c *Context), 0, 10)
			//		}
			//
			//		// global middleware
			//		ctx.handlers = append(ctx.handlers, tcpx.Mux.GlobalMiddlewares...)
			//		// anchor middleware
			//		messageIDAnchorIndex := tcpx.Mux.AnchorIndexOfMessageID(messageID)
			//		//for _, v := range tcpx.Mux.MiddlewareAnchorMap {
			//		//	if messageIDAnchorIndex > v.AnchorIndex && messageIDAnchorIndex <= v.ExpireAnchorIndex {
			//		//		ctx.handlers = append(ctx.handlers, v.Middleware)
			//		//	}
			//		//}
			//
			//		for _, v := range tcpx.Mux.MiddlewareAnchors {
			//			if messageIDAnchorIndex > v.AnchorIndex && messageIDAnchorIndex <= v.ExpireAnchorIndex {
			//				ctx.handlers = append(ctx.handlers, v.Middleware)
			//			}
			//		}
			//
			//		// self-related middleware
			//		ctx.handlers = append(ctx.handlers, tcpx.Mux.MessageIDSelfMiddleware[messageID]...)
			//		// handler
			//		ctx.handlers = append(ctx.handlers, handler)
			//
			//		if len(ctx.handlers) > 0 {
			//			ctx.Next()
			//		}
			//		ctx.Reset()
			//	}
			//}(ctx, tcpx)

			go handleMiddleware(ctx, tcpx)

			continue
		}
	}(conn, tcpx)

	select {}

	//return nil
}

func ReadAllUDP(conn net.PacketConn, maxBufferSize ...int) ([]byte, net.Addr, error) {
	if len(maxBufferSize) > 1 {
		panic(errorx.NewFromStringf("'tcpx.ListenAndServeUDP calls ReadAllUDP''s maxBufferSize should has length less by 1 but got %d", len(maxBufferSize)))
	}
	var buffer []byte
	if len(maxBufferSize) <= 0 {
		buffer = make([]byte, 4096, 4096)
	} else {
		buffer = make([]byte, maxBufferSize[0], maxBufferSize[0])
	}

	n, addr, e := conn.ReadFrom(buffer)
	fmt.Println(n)

	if e != nil {
		return nil, nil, e
	}
	return buffer[0:n], addr, nil
}

// kcp
func (tcpx *TcpX) ListenAndServeKCP(network, addr string, configs ...interface{}) error {
	listener, err := kcp.ListenWithOptions(addr, nil, 10, 3)
	defer Defer(func() {
		listener.Close()
	})
	if err != nil {
		return err
	}
	for {
		conn, e := listener.AcceptKCP()
		if e != nil {
			Logger.Println(err.Error())
			continue
		}
		ctx := NewKCPContext(conn, tcpx.Packx.Marshaller)
		if tcpx.OnConnect != nil {
			tcpx.OnConnect(ctx)
		}
		go heartBeatWatch(ctx, tcpx)

		go func(ctx *Context, tcpx *TcpX) {
			defer func() {
				if e := recover(); e != nil {
					Logger.Println(fmt.Sprintf("recover from panic %v", e))
				}
			}()
			defer ctx.UDPSession.Close()
			if tcpx.OnClose != nil {
				defer tcpx.OnClose(ctx)
			}
			var e error
			//var n int
			//var buffer = make([]byte, 1024, 1024)
			for {
				//n, e = conn.Read(buffer)
				//if e != nil {
				//	if e == io.EOF {
				//		break
				//	}
				//	fmt.Println(errorx.Wrap(e))
				//	break
				//}
				// client should send per block, rather than blocks bond together.
				// if blocks are bond, only first block are useful.
				ctx.Stream, e = tcpx.Packx.FirstBlockOf(conn)
				if e != nil {
					Logger.Println(e.Error())
					// if byte stream invalid, conn will close
					break
				}

				// Can't used prefixed by `go`
				// because requests on a same connection share context
				handleMiddleware(ctx, tcpx)

			}
		}(ctx, tcpx)
	}
	//return nil
}

// This method is abstracted from ListenAndServe[,TCP,UDP] for handling middlewares.
// When middlewares are on iterator, offset and handles are bond in 'ctx',which means when using protocol which
// shares connection/context, this function should never be used concurrently, otherwise ok.
// In specific, tcp and kcp should call like `handleMiddleware(ctx, tcpx)`, udp can call like `go handleMiddleware(ctx, tcpx)`,
// because udp meets no connection, it's no-state protocol.
//
// However, this method is not open to call everywhere.
// When rebuild new protocol server, this will be considerately used.
func handleMiddleware(ctx *Context, tcpx *TcpX) {
	if tcpx.OnMessage != nil {
		// tcpx.Mux.execAllMiddlewares(ctx)
		//tcpx.OnMessage(ctx)
		if ctx.handlers == nil {
			ctx.handlers = make([]func(c *Context), 0, 10)
		}
		ctx.handlers = append(ctx.handlers, tcpx.Mux.GlobalMiddlewares...)
		for _, v := range tcpx.Mux.MiddlewareAnchors {
			ctx.handlers = append(ctx.handlers, v.Middleware)
		}
		ctx.handlers = append(ctx.handlers, tcpx.OnMessage)
		if len(ctx.handlers) > 0 {
			ctx.Next()
		}
		ctx.Reset()
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
		if messageID == tcpx.HeartBeatMessageID && !tcpx.ThroughMiddleware {
			handler(ctx)
			return
		}

		if ctx.handlers == nil {
			ctx.handlers = make([]func(c *Context), 0, 10)
		}

		// global middleware
		ctx.handlers = append(ctx.handlers, tcpx.Mux.GlobalMiddlewares...)
		// anchor middleware
		messageIDAnchorIndex := tcpx.Mux.AnchorIndexOfMessageID(messageID)
		// ######## BUG REPORT ########
		// old: anchor type middleware may be added unordered.
		// ############################
		//for _, v := range tcpx.Mux.MiddlewareAnchorMap {
		//	if messageIDAnchorIndex > v.AnchorIndex && messageIDAnchorIndex <= v.ExpireAnchorIndex {
		//		ctx.handlers = append(ctx.handlers, v.Middleware)
		//	}
		//}
		// new:
		for _, v := range tcpx.Mux.MiddlewareAnchors {
			if messageIDAnchorIndex > v.AnchorIndex && messageIDAnchorIndex <= v.ExpireAnchorIndex {
				ctx.handlers = append(ctx.handlers, v.Middleware)
			}
		}

		// self-related middleware
		ctx.handlers = append(ctx.handlers, tcpx.Mux.MessageIDSelfMiddleware[messageID]...)
		// handler
		ctx.handlers = append(ctx.handlers, handler)

		if len(ctx.handlers) > 0 {
			ctx.Next()
		}
		ctx.Reset()
	}
}

// Start a goroutine to watch heartbeat for a connection
// When a connection is built and heartbeat mode is true, the
// then, client should do it in 5 second and continuous sends heartbeat each heart beat interval.
// ATTENTION:
// If server side set heartbeat 10s,
// client should consider the message transport price, when client send heartbeat 10s,server side might receive beyond 10s.
// Once heartbeat fail more than 3 times, it will close the connection.
func heartBeatWatch(ctx *Context, tcpx *TcpX) {
	if tcpx.HeartBeatOn == true {
		var times int
		go func() {
			for {
				select {
				case <-ctx.HeartBeatChan():
					continue
				case <-time.After(tcpx.HeatBeatInterval):
					times++
					if times == 3 {
						_ = ctx.CloseConn()
					}
					return
				}
			}
		}()
	}
}

// Before exist do ending jobs
func (tcpx TcpX) BeforeExit(f ...func()) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println(fmt.Sprintf("panic from %v", e))
			}
		}()
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
		fmt.Println("receive signal:", <-ch)
		fmt.Println("prepare to stop server")
		for _, handler := range f {
			handler()
		}
		os.Exit(0)
	}()
}
