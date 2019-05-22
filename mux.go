package tcpx

import (
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"log"
	"sync"
)

const (
	NOT_EXPIRE = 2019
)

// Mux is used to register different request by messageID
// Middlewares are divided into 3 kinds:
// 1. global  --> GlobalTypeMiddlewares
// 2. messageIDSelfRelated --> SelfRelatedTypeMiddleware
// 3. DynamicUsed --> AnchorTypeMiddleware.
// ATTENTION:
// Middlewares are executed in order of 1 ->3 -> 2
// if OnMessage is not nil, GlobalTypeMiddlewares and AnchorTypeMiddleware will all be executed regardless of unUsed or not
type Mux struct {
	Mutex    *sync.RWMutex
	Handlers map[int32]func(ctx *Context)
	AllowAdd bool

	GlobalMiddlewares       []func(ctx *Context)
	MessageIDSelfMiddleware map[int32][]func(ctx *Context)

	// expired anchors will not remove from it
	MiddlewareAnchors   []MiddlewareAnchor
	MiddlewareAnchorMap map[string]MiddlewareAnchor

	MessageIDAnchorMap map[int32]MessageIDAnchor
}

func NewMux() *Mux {
	return &Mux{
		Mutex:    &sync.RWMutex{},
		Handlers: make(map[int32]func(ctx *Context)),
		AllowAdd: true,

		GlobalMiddlewares:       make([]func(ctx *Context), 0, 10),
		MessageIDSelfMiddleware: make(map[int32][]func(ctx *Context), 0),

		MiddlewareAnchors:   make([]MiddlewareAnchor, 0, 10),
		MiddlewareAnchorMap: make(map[string]MiddlewareAnchor, 0),
		MessageIDAnchorMap:  make(map[int32]MessageIDAnchor, 0),
	}
}

// AddHandleFunc add routing handlers by messageID.
func (mux *Mux) AddHandleFunc(messageID int32, handler func(ctx *Context)) {
	if mux.AllowAdd == false {
		panic(errors.New("mux.AllowAdd is false, you should use AddHandleFunc before it's locked, after calling  tcpx.ListenAndServe(), the mux will be locked"))
	}
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()
	_, ok := mux.Handlers[messageID]
	if ok {
		panic(errors.New(fmt.Sprintf("messageID '%d' already has its handler", messageID)))
	}
	mux.Handlers[messageID] = handler
}

// anchorIndex of current handlers
func (mux *Mux) CurrentAnchorIndex() int {
	mux.Mutex.RLock()
	defer mux.Mutex.RUnlock()
	return len(mux.Handlers) - 1
}

// get anchor index of a messageID
func (mux *Mux) AnchorIndexOfMessageID(messageID int32) int {
	mux.Mutex.RLock()
	defer mux.Mutex.RUnlock()

	anchor, ok := mux.MessageIDAnchorMap[messageID]
	if !ok {
		panic(errorx.NewFromStringf("messageID '%d' anchor not found in mux.MessageIDAnchorMap", messageID))
	}
	return anchor.AnchorIndex
}

// get anchor index of a middleware
//
// Deprecated: unused in project, but it can be used in your personal test.
func (mux *Mux) AnchorIndoexOfMiddleware(middlewareKey string) (int, int) {
	mux.Mutex.RLock()
	defer mux.Mutex.RUnlock()

	anchor, ok := mux.MiddlewareAnchorMap[middlewareKey]
	if !ok {
		panic(errorx.NewFromStringf("middlewareKey '%s' anchor not found in mux.MiddlewareAnchorMap", middlewareKey))
	}
	return anchor.AnchorIndex, anchor.ExpireAnchorIndex
}

// add anchor index binding to middlewares
func (mux *Mux) AddMiddlewareAnchor(anchor MiddlewareAnchor) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()
	if mux.MiddlewareAnchorMap == nil {
		mux.MiddlewareAnchorMap = make(map[string]MiddlewareAnchor, 0)
	}
	if mux.MiddlewareAnchors == nil || len(mux.MiddlewareAnchors) == 0 {
		mux.MiddlewareAnchors = make([]MiddlewareAnchor, 0, 10)
	}
	_, ok := mux.MiddlewareAnchorMap[anchor.MiddlewareKey]
	if ok {
		panic(errorx.NewFromStringf("mux.MiddlewareAnchorMap['%s'] already exists", anchor.MiddlewareKey))
	}
	mux.MiddlewareAnchorMap[anchor.MiddlewareKey] = anchor
	mux.MiddlewareAnchors = append(mux.MiddlewareAnchors, anchor)
}

// Used to reset anchor's ExpiredAnchorIndex, avoiding operate map straightly.
func (mux *Mux) ReplaceMiddlewareAnchor(anchor MiddlewareAnchor) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()
	if mux.MiddlewareAnchorMap == nil {
		mux.MiddlewareAnchorMap = make(map[string]MiddlewareAnchor, 0)
	}
	if mux.MiddlewareAnchors == nil || len(mux.MiddlewareAnchors) == 0 {
		mux.MiddlewareAnchors = make([]MiddlewareAnchor, 0, 10)
	}
	_, ok := mux.MiddlewareAnchorMap[anchor.MiddlewareKey]
	if !ok {
		panic(errorx.NewFromStringf("mux.MiddlewareAnchorMap['%s'] not exists, can't use ReplaceMiddlewareAnchor", anchor.MiddlewareKey))
	}
	mux.MiddlewareAnchorMap[anchor.MiddlewareKey] = anchor
}

// add messageID anchor
func (mux *Mux) AddMessageIDAnchor(anchor MessageIDAnchor) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()

	if mux.MessageIDAnchorMap == nil {
		mux.MessageIDAnchorMap = make(map[int32]MessageIDAnchor, 0)
	}

	if _, ok := mux.MessageIDAnchorMap[anchor.MessageID]; ok {
		panic(errorx.NewFromStringf("mux.MessageIDAnchorMap[%d] already exists", anchor.MessageID))
	}
	mux.MessageIDAnchorMap[anchor.MessageID] = anchor
}

// add middleware by srv.Add(1, middleware1, middleware2, handler)
func (mux *Mux) AddMessageIDSelfMiddleware(messageID int32, handlers ... func(c *Context)) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()

	_, ok := mux.MessageIDSelfMiddleware[messageID]
	if ok {
		panic(errorx.NewFromStringf("messageIDSelfMiddleware[%d] already exist", messageID))
	}
	if handlers != nil && len(handlers) > 0 {
		mux.MessageIDSelfMiddleware[messageID] = make([]func(ctx *Context), 0, 10)
		mux.MessageIDSelfMiddleware[messageID] = append(mux.MessageIDSelfMiddleware[messageID], handlers...)
	}

}

func (mux *Mux) AddGlobalMiddleware(handlers ... func(c *Context)) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()
	mux.GlobalMiddlewares = append(mux.GlobalMiddlewares, handlers ...)
}

// Exec all registered middlewares.
// this function is designed to exec when tcpx.OnMessage is not nil, on this case mutex routine makes no sense, all registered
// middlewares regardless use or unUse will be exec.
//
// but,
// messageID's self-related middleware will be ignored.
//
// Deprecated: unused in project, but it can be used in your personal test.
func (mux *Mux) execAllMiddlewares(ctx *Context) {
	for _, handler := range mux.GlobalMiddlewares {
		handler(ctx)
		if ctx.offset == ABORT {
			return
		}
	}
	ctx.ResetOffset()
	for key, middlewareAnchor := range mux.MiddlewareAnchors {
		log.Println(key)
		middlewareAnchor.Middleware(ctx)
		if ctx.offset == ABORT {
			return
		}
	}
	ctx.ResetOffset()
}

// exec middlewares added by srv.Add(1, middleware1, middleware2, handler)
//
// Deprecated: unused in project, but it can be used in your personal test.
func (mux *Mux) execMessageIDMiddlewares(ctx *Context, messageID int32) {
	for _, handler := range mux.GlobalMiddlewares {
		handler(ctx)
		if ctx.offset == ABORT {
			return
		}
	}
	ctx.ResetOffset()
	var middlewareAnchorIndex, messagIDAnchorIndex int
	var middlewareExpireAnchorIndex int
	for k, middlewareAnchor := range mux.MiddlewareAnchorMap {
		middlewareAnchorIndex, middlewareExpireAnchorIndex = mux.AnchorIndoexOfMiddleware(k)
		messagIDAnchorIndex = mux.AnchorIndexOfMessageID(messageID)

		if messagIDAnchorIndex > middlewareAnchorIndex && messagIDAnchorIndex <= middlewareExpireAnchorIndex {
			middlewareAnchor.Middleware(ctx)
			if ctx.offset == ABORT {
				return
			}
		}
	}
	ctx.ResetOffset()

	var selfMiddlewares []func(ctx *Context)
	var ok bool
	if selfMiddlewares, ok = mux.MessageIDSelfMiddleware[messageID]; ok {
		if selfMiddlewares != nil && len(selfMiddlewares) > 0 {
			for _, handler := range selfMiddlewares {
				handler(ctx)
				if ctx.offset == ABORT {
					return
				}
			}
		}
	}
	ctx.ResetOffset()

}

// exec all global middlewares
//
// Deprecated: unused in project, but it can be used in your personal test.
func (mux *Mux) execGlobalMiddlewares(ctx *Context) {
	for _, handler := range mux.GlobalMiddlewares {
		handler(ctx)
		if ctx.offset == ABORT {
			return
		}
	}
	ctx.ResetOffset()
}

