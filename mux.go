package tcpx

import (
	"errorX"
	"errors"
	"fmt"
	"sync"
)

const(
	NOT_EXPIRE = 2019
)
// mux is used to register different request by messageID
// middlewares are divided into 3 kinds:
// 1. global  --> GlobalMiddlewares
// 2. messageIDSelfRelated --> MessageIDSelfMiddleware
// 3. DynamicUsed --> MiddlewareAnchors.
// ATTENTION:
// middlewares are executed in order of 1 ->3 -> 2
type Mux struct {
	Mutex    *sync.RWMutex
	Handlers map[int32]func(ctx *Context)
	AllowAdd bool

	GlobalMiddlewares       []func(ctx *Context)
	MessageIDSelfMiddleware map[int32][]func(ctx *Context)

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
func (mux *Mux) AnchorIndoexOfMiddleware(middlewareKey string) (int,int) {
	mux.Mutex.RLock()
	defer mux.Mutex.RUnlock()

	anchor, ok := mux.MiddlewareAnchorMap[middlewareKey]
	if !ok {
		panic(errorx.NewFromStringf("middlewareKey '%s' anchor not found in mux.MiddlewareAnchorMap", middlewareKey))
	}
	return anchor.AnchorIndex,anchor.ExpireAnchorIndex
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
	mux.MiddlewareAnchors = append(mux.MiddlewareAnchors, anchor)
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

// Exec all registered middlewares.
// this function is designed to exec when tcpx.OnMessage is not nil, on this case mutex routine makes no sense, all registered
// middlewares regardless use or unUse will be exec.
//
// but,
// messageID's self-related middleware will be ignored
func (mux *Mux) execAllMiddlewares(ctx *Context) {
	for _, handler := range mux.GlobalMiddlewares {
		handler(ctx)
		if ctx.offset == ABORT {
			return
		}
	}
	ctx.ResetOffset()
	for _, middlewareAnchor := range mux.MiddlewareAnchors {
		middlewareAnchor.Middleware(ctx)
		if ctx.offset == ABORT {
			return
		}
	}
	ctx.ResetOffset()
}

func (mux *Mux) execMessageIDMiddlewares(ctx *Context, messageID int32) {
	for _, handler := range mux.GlobalMiddlewares {
		handler(ctx)
		if ctx.offset == ABORT {
			return
		}
	}
	ctx.ResetOffset()
	var middlewareAnchorIndex,middlewareExpireAnchorIndex, messagIDAnchorIndex int
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

func (mux *Mux) AddMessageIDSelfMiddleware(messageID int32, handlers ... func(c *Context)) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()

	_, ok :=mux.MessageIDSelfMiddleware[messageID]
	if ok{
		panic(errorx.NewFromStringf("messageIDSelfMiddleware[%d] already exist", messageID))
	}
	if handlers ==nil {
		mux.MessageIDSelfMiddleware[messageID] = make([]func(ctx *Context), 0,10)
		mux.MessageIDSelfMiddleware[messageID] = append(mux.MessageIDSelfMiddleware[messageID], handlers...)
	}

}
