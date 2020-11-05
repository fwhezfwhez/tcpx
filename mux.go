package tcpx

import (
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"sync"
	"sync/atomic"
)

const (
	// context's anchor middleware will expire when call UnUse(),
	// middleware added by Use() will be set 2019 anchor index by default
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
	indexSeed int32
	// mux instance lock
	Mutex *sync.RWMutex

	// handlers of messageID routers
	Handlers map[int32]func(ctx *Context)

	AllowAdd bool

	// global-middlewares
	GlobalMiddlewares []func(ctx *Context)
	// messageID middlewares
	MessageIDSelfMiddleware map[int32][]func(ctx *Context)

	// all middleware anchors, expired anchors will not remove from it
	MiddlewareAnchors []MiddlewareAnchor
	// all middleware anchors
	MiddlewareAnchorMap map[string]MiddlewareAnchor
	// messageID handlers anchors
	MessageIDAnchorMap map[int32]MessageIDAnchor

	// urlMux
	urlMux *URLMux
}

// New a mux instance, malloc memory for its mutex, handler slice...
func NewMux() *Mux {
	return &Mux{
		indexSeed: 1,
		Mutex:     &sync.RWMutex{},
		Handlers:  make(map[int32]func(ctx *Context)),
		AllowAdd:  true,

		GlobalMiddlewares:       make([]func(ctx *Context), 0, 10),
		MessageIDSelfMiddleware: make(map[int32][]func(ctx *Context), 0),

		MiddlewareAnchors:   make([]MiddlewareAnchor, 0, 10),
		MiddlewareAnchorMap: make(map[string]MiddlewareAnchor, 0),
		MessageIDAnchorMap:  make(map[int32]MessageIDAnchor, 0),


		urlMux: NewURLMux(),
	}
}

// Any is used to routing message using url-pattern
func (mux *Mux) Any(urlPattern string, handlers ... func(c *Context)) error {
	if mux.isReadOnly() == false {
		if e := mux.urlMux.AddURLPatternHandler(urlPattern, handlers...); e != nil {
			return errorx.Wrap(e)
		}
		return nil
	} else {
		return errorx.NewFromString("mux is only writable before mux.LockWrite()")
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
	return int(atomic.AddInt32(&mux.indexSeed, 1))
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

func (mux *Mux) AnchorIndexOfURLPattern(urlPattern string) int {
	mux.Mutex.RLock()
	defer mux.Mutex.RUnlock()

	anchor, ok := mux.urlMux.URLAnchorMap[urlPattern]
	if !ok {
		panic(errorx.NewFromStringf("urlPattern '%s' anchor not found in mux.urlMux.URLPatternAnchorMap", urlPattern))
	}
	return anchor.AnchorIndex
}

// get anchor index of a middleware
//
// Deprecated: unused in project, but it can be used in your personal test.
//func (mux *Mux) AnchorIndexOfMiddleware(middlewareKey string) (int, int) {
//	mux.Mutex.RLock()
//	defer mux.Mutex.RUnlock()
//
//	anchor, ok := mux.MiddlewareAnchorMap[middlewareKey]
//	if !ok {
//		panic(errorx.NewFromStringf("middlewareKey '%s' anchor not found in mux.MiddlewareAnchorMap", middlewareKey))
//	}
//	return anchor.AnchorIndex, anchor.ExpireAnchorIndex
//}

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

		return
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
	// replace map
	mux.MiddlewareAnchorMap[anchor.MiddlewareKey] = anchor

	// replace slice
L:
	for i, v := range mux.MiddlewareAnchors {
		if v.MiddlewareKey == anchor.MiddlewareKey {
			mux.MiddlewareAnchors[i] = anchor
			break L
		}
	}
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

// add url-pattern anchor
func (mux *Mux) AddURLAnchor(anchor MessageIDAnchor) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()

	if mux.urlMux.URLAnchorMap == nil {
		mux.urlMux.URLAnchorMap = make(map[string]MessageIDAnchor, 0)
	}

	if _, ok := mux.urlMux.URLAnchorMap[anchor.URLPattern]; ok {
		panic(errorx.NewFromStringf("mux.urlMux.URLAnchorMap[%s] already exists", anchor.URLPattern))
	}
	mux.urlMux.URLAnchorMap[anchor.URLPattern] = anchor
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

// Add Global middlewares
func (mux *Mux) AddGlobalMiddleware(handlers ... func(c *Context)) {
	mux.Mutex.Lock()
	defer mux.Mutex.Unlock()
	mux.GlobalMiddlewares = append(mux.GlobalMiddlewares, handlers ...)
}

func (mux *Mux) isReadOnly() bool {
	return !mux.AllowAdd
}

// Exec all registered middlewares.
// this function is designed to exec when tcpx.OnMessage is not nil, on this case mutex routine makes no sense, all registered
// middlewares regardless use or unUse will be exec.
//
// but,
// messageID's self-related middleware will be ignored.
//
// Deprecated: unused in project, but it can be used in your personal test.
//func (mux *Mux) execAllMiddlewares(ctx *Context) {
//	for _, handler := range mux.GlobalMiddlewares {
//		handler(ctx)
//		if ctx.offset == ABORT {
//			return
//		}
//	}
//	ctx.ResetOffset()
//	for key, middlewareAnchor := range mux.MiddlewareAnchors {
//		log.Println(key)
//		middlewareAnchor.Middleware(ctx)
//		if ctx.offset == ABORT {
//			return
//		}
//	}
//	ctx.ResetOffset()
//}
//
//// exec middlewares added by srv.Add(1, middleware1, middleware2, handler)
////
//// Deprecated: unused in project, but it can be used in your personal test.
//func (mux *Mux) execMessageIDMiddlewares(ctx *Context, messageID int32) {
//	for _, handler := range mux.GlobalMiddlewares {
//		handler(ctx)
//		if ctx.offset == ABORT {
//			return
//		}
//	}
//	ctx.ResetOffset()
//	var middlewareAnchorIndex, messagIDAnchorIndex int
//	var middlewareExpireAnchorIndex int
//	for k, middlewareAnchor := range mux.MiddlewareAnchorMap {
//		middlewareAnchorIndex, middlewareExpireAnchorIndex = mux.AnchorIndexOfMiddleware(k)
//		messagIDAnchorIndex = mux.AnchorIndexOfMessageID(messageID)
//
//		if messagIDAnchorIndex > middlewareAnchorIndex && messagIDAnchorIndex <= middlewareExpireAnchorIndex {
//			middlewareAnchor.Middleware(ctx)
//			if ctx.offset == ABORT {
//				return
//			}
//		}
//	}
//	ctx.ResetOffset()
//
//	var selfMiddlewares []func(ctx *Context)
//	var ok bool
//	if selfMiddlewares, ok = mux.MessageIDSelfMiddleware[messageID]; ok {
//		if selfMiddlewares != nil && len(selfMiddlewares) > 0 {
//			for _, handler := range selfMiddlewares {
//				handler(ctx)
//				if ctx.offset == ABORT {
//					return
//				}
//			}
//		}
//	}
//	ctx.ResetOffset()
//
//}
//
//// exec all global middlewares
////
//// Deprecated: unused in project, but it can be used in your personal test.
//func (mux *Mux) execGlobalMiddlewares(ctx *Context) {
//	for _, handler := range mux.GlobalMiddlewares {
//		handler(ctx)
//		if ctx.offset == ABORT {
//			return
//		}
//	}
//	ctx.ResetOffset()
//}
//
