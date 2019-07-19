// Package tcpx provides udp,tcp,kcp three kinds of protocol.
package tcpx

import (
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"github.com/xtaci/kcp-go"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"net"
)

const (
	DEFAULT_HEARTBEAT_MESSAGEID = 1392
	STATE_RUNNING               = 1
	STATE_STOP                  = 2
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

	// built-in clientPool
	// clientPool is defined in github.com/tcpx/clientPool/client-pool.go, you might design your own pool yourself as
	// long as you set builtInPool = false
	// - How to add/delete an connection into/from pool?
	// ```
	//    // add
	//    ctx.Online(username)
	//    // delete
	//    ctx.Offline(username)
	// ``
	builtInPool bool
	pool        *ClientPool

	// external for restart
	properties []*PropertyCache
	pLock      *sync.RWMutex
	state      int // 1- running, 2- stopped

	// broadcast some signal to all connection
	withSignals    bool
	closeAllSignal chan int // used to close all connection
}

type PropertyCache struct {
	Network string
	Port    string

	// only when network is 'tcp','kcp', Listener can assert to net.Listener.
	// when network is 'udp', it can assert to net.PackConn
	Listener interface{}
}

// new an tcpx srv instance
func NewTcpX(marshaller Marshaller) *TcpX {
	return &TcpX{
		Packx:      NewPackx(marshaller),
		Mux:        NewMux(),
		properties: make([]*PropertyCache, 0, 10),
		pLock:      &sync.RWMutex{},
		state:      2,
	}
}

// whether using built-in pool
func (tcpx *TcpX) WithBuiltInPool(yes bool) *TcpX {
	tcpx.builtInPool = yes
	tcpx.pool = NewClientPool()
	return tcpx
}

// Whether using signal-broadcast.
// Used for these situations:
// closeAllSignal - close all connection and remove them from the built-in pool
func (tcpx *TcpX) WithBroadCastSignal(yes bool) *TcpX {
	tcpx.withSignals = yes
	tcpx.closeAllSignal = make(chan int, 1)
	return tcpx
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
// * If you want specific official heartbeat handler detail:
// srv.HeartBeatModeDetail(true, 10 * time.Second, true, 1)
//
// * If you want to rewrite heartbeat handler:
// srv.RewriteHeartBeatHandler(func(c *tcpx.Context){})
//
// * If you think built in heartbeat not good, abandon it:
// ```
// srv.AddHandler(1111, func(c *tcpx.Context){
//    //do nothing by default and define your heartbeat yourself
// })
// ```
func (tcpx *TcpX) HeartBeatMode(on bool, duration time.Duration) *TcpX {
	tcpx.HeartBeatOn = on
	tcpx.HeatBeatInterval = duration
	tcpx.ThroughMiddleware = false
	tcpx.HeartBeatMessageID = DEFAULT_HEARTBEAT_MESSAGEID

	if on {
		tcpx.AddHandler(DEFAULT_HEARTBEAT_MESSAGEID, func(c *Context) {
			Logger.Println(fmt.Sprintf("recv '%s' heartbeat:", c.ClientIP()), c.Stream)
			c.RecvHeartBeat()
		})
	}
	return tcpx
}

// specific args for heartbeat
func (tcpx *TcpX) HeartBeatModeDetail(on bool, duration time.Duration, throughMiddleware bool, messageID int32) *TcpX {
	tcpx.HeartBeatOn = on
	tcpx.HeatBeatInterval = duration
	tcpx.ThroughMiddleware = throughMiddleware
	tcpx.HeartBeatMessageID = messageID

	if on {
		tcpx.AddHandler(messageID, func(c *Context) {
			Logger.Println(fmt.Sprintf("recv '%s' heartbeat:", c.ClientIP()), c.Stream)
			c.RecvHeartBeat()
		})
	}
	return tcpx
}

// Rewrite heartbeat handler
// It will inherit properties of the older heartbeat handler:
//   * heartbeatInterval
//   * throughMiddleware
func (tcpx *TcpX) RewriteHeartBeatHandler(messageID int32, f func(c *Context)) *TcpX {
	tcpx.removeHandler(tcpx.HeartBeatMessageID)
	tcpx.HeartBeatMessageID = messageID
	tcpx.AddHandler(messageID, f)
	return tcpx
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
	//if In(network, []string{"http", "https"}) {
	//	return tcpx.ListenAndServeHTTP(network, addr)
	//}
	return errorx.NewFromStringf("'network' doesn't support '%s'", network)
}

func (tcpx *TcpX) fillProperty(network, addr string, listener interface{}) {

	tcpx.pLock.Lock()
	defer tcpx.pLock.Unlock()

	if tcpx.properties == nil {
		tcpx.properties = make([]*PropertyCache, 0, 10)
	}

	// if property exists, only replace listener
	for i, v := range tcpx.properties {
		if v.Network == network && v.Port == addr {
			tcpx.properties[i].Listener = listener
			return
		}
	}
	prop := &PropertyCache{
		Network:  network,
		Port:     addr,
		Listener: listener,
	}
	tcpx.properties = append(tcpx.properties, prop)
}

// tcp
func (tcpx *TcpX) ListenAndServeTCP(network, addr string) error {
	defer func() {
		if e := recover(); e != nil {
			Logger.Println(fmt.Sprintf("recover from panic %v", e))
			Logger.Println(string(debug.Stack()))
			return
		}
	}()
	listener, err := net.Listen(network, addr)
	fmt.Println(2, "new listener")
	if err != nil {
		return err
	}
	tcpx.fillProperty(network, addr, listener)

	defer listener.Close()
	tcpx.openState()
	for {

		if tcpx.State() == STATE_STOP {
			break
		}
		conn, err := listener.Accept()
		if err != nil {
			log.Println(fmt.Sprintf(err.Error()))
			break
		}
		ctx := NewContext(conn, tcpx.Packx.Marshaller)

		if tcpx.builtInPool {
			ctx.poolRef = tcpx.pool
		}

		if tcpx.OnConnect != nil {
			tcpx.OnConnect(ctx)
		}

		go broadcastSignalWatch(ctx, tcpx)
		go heartBeatWatch(ctx, tcpx)

		go func(ctx *Context, tcpx *TcpX) {
			defer func() {
				if e := recover(); e != nil {
					Logger.Println(fmt.Sprintf("recover from panic %v", e))
					Logger.Println(string(debug.Stack()))
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

				handleMiddleware(ctx, tcpx)
				continue
			}
		}(ctx, tcpx)
	}
	return nil
}

// set srv state running
func (tcpx *TcpX) openState() {
	tcpx.pLock.Lock()
	defer tcpx.pLock.Unlock()
	tcpx.state = 1
}

// set srv state stopped
// tcp, udp, kcp will stop for circle and close listener/conn
func (tcpx *TcpX) stopState() {
	tcpx.pLock.Lock()
	defer tcpx.pLock.Unlock()
	tcpx.state = STATE_STOP
}
func (tcpx *TcpX) State() int {
	tcpx.pLock.RLock()
	defer tcpx.pLock.RUnlock()
	return tcpx.state
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
	defer conn.Close()

	tcpx.fillProperty(network, addr, conn)

	tcpx.openState()
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
			if tcpx.State() == STATE_STOP {
				break
			}
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

			go broadcastSignalWatch(ctx, tcpx)

			go heartBeatWatch(ctx, tcpx)
			if tcpx.builtInPool {
				ctx.poolRef = tcpx.pool
			}

			ctx.Stream, e = tcpx.Packx.FirstBlockOfBytes(buffer)
			if e != nil {
				Logger.Println(e.Error())
				continue
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
// all configs are using default value.
func (tcpx *TcpX) ListenAndServeKCP(network, addr string, configs ...interface{}) error {
	listener, err := kcp.ListenWithOptions(addr, nil, 10, 3)
	//defer Defer(func() {
	//	listener.Close()
	//})
	if err != nil {
		return err
	}
	defer listener.Close()

	tcpx.fillProperty(network, addr, listener)

	tcpx.openState()
	for {
		if tcpx.State() == STATE_STOP {
			break
		}
		conn, e := listener.AcceptKCP()
		if e != nil {
			Logger.Println(err.Error())
			continue
		}
		ctx := NewKCPContext(conn, tcpx.Packx.Marshaller)

		if tcpx.builtInPool {
			ctx.poolRef = tcpx.pool
		}

		if tcpx.OnConnect != nil {
			tcpx.OnConnect(ctx)
		}
		// signal management
		go broadcastSignalWatch(ctx, tcpx)

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
	return nil
}

// http
// developing, do not use.
//
// Deprecated: on developing.
func (tcpx *TcpX) ListenAndServeHTTP(network, addr string) error {
	r := gin.New()
	r.Any("/tcpx/message/:messageID/", func(ginCtx *gin.Context) {

	})
	s := &http.Server{
		Addr:           addr,
		Handler:        cors.AllowAll().Handler(r),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 21,
	}
	return s.ListenAndServe()
}

// grpc
// developing, do not use.
// marshaller must be protobuf, clients should send message bytes which body is protobuf bytes
//
// Deprecated: on developing.
func (tcpx *TcpX) ListenAndServeGRPC(network, addr string) error {
	return nil
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
// In these cases heartbeat watching goroutine will stop:
// - tcpx.closeAllSignal: when tcpx srv calls `srv.Stop()`, closeAllSignal will be closed and stop this watching goroutine.
// - ctx.recvEnd: when connection's context calls 'ctx.CloseConn()', recvEnd will be closed and stop this watching goroutine.
// - time out receiving interval heartbeat pack.
func heartBeatWatch(ctx *Context, tcpx *TcpX) {
	if tcpx.HeartBeatOn == true {
		var times int
		go func() {
			for {
				if tcpx.State() == STATE_STOP {
					ctx.CloseConn()
					break
				}
				select {
				case <-ctx.HeartBeatChan():
					continue
				case <-time.After(tcpx.HeatBeatInterval):
					times++
					if times == 3 {
						_ = ctx.CloseConn()
					}
					return
				case <-tcpx.closeAllSignal:
					ctx.CloseConn()
					return
				case <-ctx.recvEnd:
					return
				}
			}
		}()
	}
}

// Each connection will have this goroutine, it bind relation with tcpx server.
// tcpx.closeAllSignal: when tcpx srv calls `srv.Stop()`, closeAllSignal will be closed and stop this watching goroutine.
// ctx.recvEnd: when connection's context calls 'ctx.CloseConn()', recvEnd will be closed and stop this watching goroutine.
func broadcastSignalWatch(ctx *Context, tcpx *TcpX) {
	if tcpx.withSignals == true {
		for {
			select {
			case <-tcpx.closeAllSignal:
				ctx.CloseConn()
				return
			case <-ctx.recvEnd:
				return
			}
		}
	}
}

func online(ctx *Context, tcpx *TcpX) {

}
func offline(ctx *Context, tcpx *TcpX) {

}

// Before exist do ending jobs
func (tcpx *TcpX) BeforeExit(f ...func()) {
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

// Graceful stop server parts generated by `srv.ListenAndServe()`, this will not stop process, if param 'closeAllConnection' is false, only stop server listener.
// Older connections will remain safe and kept in pool.If param 'closeAllConnection' is true, it will not only stop the
// listener, but also kill all connections(stops their net.Conn, stop all sub-routine, clear the pool)
func (tcpx *TcpX) Stop(closeAllConnection bool) error {
     fmt.Println("graceful stop")
	if tcpx.State() == STATE_STOP {
		return errors.New("already stopped")
	}

	tcpx.stopState()

	// close all listener
	func() {
		tcpx.pLock.Lock()
		defer tcpx.pLock.Unlock()
		for i, v := range tcpx.properties {
			switch v.Network {
			case "kcp", "tcp":
				fmt.Println(1, "closing")
				tcpx.properties[i].Listener.(net.Listener).Close()
			case "udp":
				tcpx.properties[i].Listener.(net.PacketConn).Close()
			}
		}
	}()

	// close all connections
	if closeAllConnection == true {
		tcpx.closeAllConnection()
	}
	return nil
}

func (tcpx *TcpX) closeAllConnection() {
	if tcpx.withSignals == true {
		close(tcpx.closeAllSignal)
	} else {
		if tcpx.pool != nil {
			oldPool := tcpx.pool
			go func() {
				oldPool.m.Lock()
				defer oldPool.m.Unlock()

				for k, _ := range oldPool.Clients {
					oldPool.Clients[k].CloseConn()
					delete(oldPool.Clients, k)
				}
			}()

			tcpx.pool = NewClientPool()
		}
	}
}

// Graceful start an existed tcpx srv, former server is stopped by tcpX.Stop()
func (tcpx *TcpX) Start() error {
	if tcpx.State() == STATE_RUNNING {
		return errors.New("already running")
	}

	for _, v := range tcpx.properties {
		go func() {
			defer func() {
				if e := recover(); e != nil {
					Logger.Println(fmt.Sprintf("panic from '%v' \n %s", e, debug.Stack()))
				}
			}()
			fmt.Println(fmt.Sprintf("graceful restart %s server on %s", v.Network, v.Port))
			e := tcpx.ListenAndServe(v.Network, v.Port)
			if e != nil {
				Logger.Println(fmt.Sprintf("%s \n %s", e.Error(), debug.Stack()))
			}
		}()
	}
	return nil
}

// Graceful Restart = Stop and Start.Besides, you can
func (tcpx *TcpX) Restart(closeAllConnection bool, beforeStart ... func()) error {
	if e:=tcpx.Stop(closeAllConnection);e!=nil {
		return e
	}

	for _, v := range beforeStart {
		v()
	}
	// time.Sleep(5 * time.Second)
	if e:=tcpx.Start();e!=nil{
		return e
	}
	return nil
}
