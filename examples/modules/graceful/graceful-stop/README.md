## graceful-stop

#### detail
There is four situation:
- **Has set closeAll broadcast signal**
After calling `srv.WithBroadCastSignal(true)`, all coming-in connections will get control. when calls `srv.Stop()`, all listener and connection will finally clear.

- ** without closeAll broadcast signal but has built-in pool **
After calling `srv.WithBuiltInPool(true)`, all online connection will get control, closed after 10s. But if connection not call `ctx.Online()`, this connection will not be stopped.It will leak until its deadline or client break it, or close api.

- **without signal and pool**
After 10s, server side only refuse server, but old connections is still working on until its deadline or break by client or close api.

#### step
`go run main.go`

`cd client`

`go run client.go`

#### result
After 10 seconds, server side receives log listener closed, client side receives EOF.
