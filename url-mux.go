package tcpx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"runtime"
)

// ## introduction:
// url-mux is another routing style for tcpx. It works like:
/*
    srv.Any("/bra/bra", func(c *tcpx.Context){})
*/
// All usage about url-mux is alike messageID mux.

// ## Why to design url-mux?
// Since messageID style requires users to manage messageID themselves, it's kind of inconvenient for a team of many developers to manage conflicted messageIDs.
// Also, url-style is better read.

type URLMux struct {
	// messageIDMux  map[int][]func(c *Context)
	urlPatternMux map[string][]func(c *Context)
	URLAnchorMap  map[string]MessageIDAnchor

	readOnly bool

	urlRouteInfo map[string]Route

	panicOnExistRouting bool
}

func NewURLMux() *URLMux {
	return &URLMux{
		urlPatternMux: make(map[string][]func(c *Context)),

		urlRouteInfo: make(map[string]Route),
	}
}

// 基于url-pattern添加路由
func (m *URLMux) AddURLPatternHandler(urlPattern string, handlers ... func(c *Context)) error {
	if m.readOnly == false {
		_, file, line, _ := runtime.Caller(1)
		routeInfo := Route{
			Whereis:    []string{fmt.Sprintf("%s:%d", file, line)},
			URLPattern: urlPattern,
		}

		h, ok := m.urlPatternMux[urlPattern]
		if ok {
			if m.panicOnExistRouting {
				panic(fmt.Errorf("handler conflicts on the same url-pattern: \n%s\nThe existed route-info is at:\n%s", routeInfo.Whereis[0], m.urlRouteInfo[urlPattern].Location()))
			}

			m.urlPatternMux[urlPattern] = append(h, handlers...)
		} else {
			m.urlPatternMux[urlPattern] = handlers
		}

		r, exist := m.urlRouteInfo[urlPattern]
		if !exist {
			m.urlRouteInfo[urlPattern] = routeInfo
		} else {
			m.urlRouteInfo[urlPattern] = r.Merge(routeInfo)
		}
		return nil
	} else {
		return errorx.NewFromString("mux is only writable before mux.LockWrite()")
	}
}

// MessageID和URL路由在添加时，如果已存在，则会panic。
func (m *URLMux) PanicOnExistRouter() error {
	if m.readOnly == false {
		m.panicOnExistRouting = true
		return nil
	} else {
		return errorx.NewFromString("mux is only writable before mux.LockWrite()")
	}
}

// 锁定后，无法再添加路由
func (m *URLMux) LockWrite() {
	m.readOnly = true
}
