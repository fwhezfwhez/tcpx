package tcpx

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/fwhezfwhez/cmap"
	"github.com/fwhezfwhez/errorx"
)

// Call args
type Option struct {
	Cache *cmap.Map
	l     *sync.RWMutex

	Network string
	Host    string

	Timeout    time.Duration
	Marshaller Marshaller

	KeepAlive bool
	AliveTime time.Duration
}

func NewOption() *Option {
	return &Option{
		Cache: cmap.NewMap(),
		l:     &sync.RWMutex{},
	}
}

func (o *Option) SetNetworkHost(network, host string) *Option {
	o.Network = network
	o.Host = host
	return o
}
func (o *Option) Option(option Option) *Option {
	if option.Host != "" {
		o.Host = option.Host
	}
	if option.Network != "" {
		o.Network = option.Network
	}
	if option.Timeout != 0 {
		o.Timeout = option.Timeout
	}

	if option.KeepAlive != false {
		o.KeepAlive = option.KeepAlive
	}
	if option.Marshaller != nil {
		o.Marshaller = option.Marshaller
	}

	if option.AliveTime != 0 {
		o.AliveTime = option.AliveTime
	}
	return o
}

// Copy an option instance for different request
func (o *Option) Copy() *Option {
	return &Option{
		Network: o.Network,
		Host:    o.Host,
		Cache:   o.Cache,
		Timeout: o.Timeout,

		Marshaller: o.Marshaller,

		KeepAlive: o.KeepAlive,
		AliveTime: o.AliveTime,
	}
}

func (o *Option) setConn(conn net.Conn, connHash string) {
	o.l.Lock()
	defer o.l.Unlock()

	o.Cache.Set(connHash, conn)
}

func (o *Option) getConn(network string, host string) (net.Conn, error) {
	var conn net.Conn
	var e error

	connHash := connHash(network, host)

	o.l.RLock()
	connI, ok := o.Cache.Get(connHash)
	o.l.RUnlock()

	if !ok {
		conn, e = net.Dial(network, host)
		if e != nil {
			return nil, errorx.Wrap(e)
		}
		o.setConn(conn, connHash)
		return conn, nil
	} else {
		conn = connI.(net.Conn)
	}

	return conn, nil
}

// Call require client send a request and server response once to once
func call(request []byte, host string, network string, option *Option) ([]byte, error) {
	var result = make(chan []byte, 1)
	var connHash = connHash(network, host)

	conn, e := option.getConn(network, host)
	if e != nil {
		return nil, errorx.Wrap(e)
	}

	if option.KeepAlive == true {
		go func() {
			for {
				select {
				case <-time.After(option.AliveTime):
					conn.Close()
					option.l.Lock()
					option.Cache.Delete(connHash)
					option.l.Unlock()
				}
			}
		}()
	}
	go func() {
		var block []byte
		var e error
		block, e = FirstBlockOf(conn)
		if e != nil {
			log.Println(e)
			return
		}
		result <- block
		return
	}()
	conn.Write(request)
	if option.Timeout == 0 {
		v := <-result
		return v, nil
	} else {
		select {
		case <-time.After(option.Timeout):
			return nil, fmt.Errorf("time out")
		case v := <-result:
			return v, nil
		}
	}
}

func connHash(network string, host string) string {
	return fmt.Sprintf("%s://%s", network, host)
}

func connExpHash(connHash string) string {
	return fmt.Sprintf("%s/%s", connHash, "exp")
}
