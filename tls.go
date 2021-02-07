package tcpx

import (
	"crypto/tls"
	"github.com/fwhezfwhez/errorx"
	"net"
)

type listenerConfig struct {
	Network   string
	Addr      string
	TLSConfig *tls.Config
}

func newListener(lc listenerConfig) (net.Listener, error) {
	if lc.TLSConfig == nil {
		ls, e := net.Listen(lc.Network, lc.Addr)
		if e != nil {
			return nil, errorx.Wrap(e)
		}
		return ls, nil
	}

	return tls.Listen(lc.Network, lc.Addr, lc.TLSConfig)
}
