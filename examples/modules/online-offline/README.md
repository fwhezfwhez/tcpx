## online-offline
This example shows how to use official built-in pool to offline/online.

#### detail
Tcpx provides an official clientPool to manager client contexts. Here I show how to use it online and offline

#### step
`cd server`

`go run server.go`

`cd client`

`go run client.go`

#### result
server-side output 'online success' and after 5 seconds output 'offline success'.
