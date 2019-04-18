package tcpx

import (
	"errors"
	"fmt"
	"sync"
)

// mux is used to register different request by messageID
type Mux struct {
	Mutex    *sync.RWMutex
	Handlers map[int32]func(ctx *Context)
	AllowAdd bool
}

func NewMux() *Mux {
	return &Mux{
		Mutex:    &sync.RWMutex{},
		Handlers: make(map[int32]func(ctx *Context)),
		AllowAdd: true,
	}
}

// AddHandleFunc add routing handlers by messageID.
// AddHandleFunc is not concurrently safe, don't use it in concurrent goroutines like:
// go func(){
//     mux.AddHandleFunc(1,f1)
// }
// go func(){
//     mux.AddHandleFunc(2,f2)
// }
func (mux *Mux)AddHandleFunc(messageID int32, handler func(ctx *Context)){
	if mux.AllowAdd == false{
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
