## Auth
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/auth

Auth makes different sense comparing with middleware. A middleware can easily stop a invalid request after a connection has been established, but It can't avoid a client keep sending heartbeat but do nothing.It still occupy a connection resource.

Auth will start a goroutine once a connection is on. In a specific interval not receiving signal, connection will be forcely dropped by server side.

server.go
```go
srv.WithAuthDetail(true, 30*time.Second, false, tcpx.DEFAULT_AUTH_MESSAGEID, func(c *tcpx.Context) {
		c.RecvAuthPass()
})
```
