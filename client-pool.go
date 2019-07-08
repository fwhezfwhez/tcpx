package tcpx

import (
	"sync"
)

var GlobalClientPool *ClientPool

type ClientPool struct {
	Clients map[string]*Context
	m       *sync.RWMutex
}


func NewClientPool() *ClientPool {
	return &ClientPool{
		Clients: make(map[string]*Context),
		m:       &sync.RWMutex{},
	}
}
func init() {
	GlobalClientPool = NewClientPool()
}
func (cp *ClientPool) SetClientPool(username string, ctx *Context) {
	cp.m.Lock()
	defer cp.m.Unlock()
	cp.Clients[username] = ctx
}

func (cp *ClientPool) GetClientPool(username string) *Context {
	cp.m.RLock()
	defer cp.m.RUnlock()
	return cp.Clients[username]
}

func (cp *ClientPool) DeleteFromClientPool(username string) {
	cp.m.RLock()
	defer cp.m.RUnlock()
	delete(cp.Clients, username)
}

func (cp *ClientPool) Online(username string, ctx *Context) {
	cp.SetClientPool(username, ctx)
}
func (cp *ClientPool) Offline(username string) {
	ctx := cp.GetClientPool(username)
	if ctx != nil {
		ctx.CloseConn()
		cp.DeleteFromClientPool(username)
	}
}
