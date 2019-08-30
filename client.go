package tcpx

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Call args
type Option struct {
	Cache map[string]net.Conn
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
		Cache: make(map[string]net.Conn, 0),
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

// Call require client send a request and server response once to once
func Call(request []byte, option *Option) ([]byte, error) {
	var result = make(chan []byte, 1)
	var connHash = fmt.Sprintf("%s://%s", option.Network, option.Host)
	var e error
	var conn net.Conn

	// get conn from pool
	if len(option.Cache) != 0 {
		var ok bool
		option.l.RLock()
		conn, ok = option.Cache[connHash]
		option.l.RUnlock()
		if !ok {
			conn, e = net.Dial(option.Network, option.Host)
			if e != nil {
				return nil, e
			}
			option.l.Lock()
			option.Cache[connHash] = conn
			option.l.Unlock()
		}
	} else {
		conn, e = net.Dial(option.Network, option.Host)
		if e != nil {
			return nil, e
		}
		option.l.Lock()
		option.Cache[connHash] = conn
		option.l.Unlock()

	}
	if e != nil {
		return nil, e
	}

	if option.KeepAlive == true {
		go func() {
			for {
				select {
				case <-time.After(option.AliveTime):
					conn.Close()
					option.l.Lock()
					delete(option.Cache, connHash)
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
