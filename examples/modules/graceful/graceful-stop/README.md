## graceful-stop

#### step
`go run main.go`

`cd client`

`go run client.go` **normal**

after 10s
`go run client.go` **refused**

#### result
if `main.go` call `srv.Stop(false)`:
After 10 seconds, server side receives log listener closed, old client connections remain but new client are refused.

if `main.go` call `srv.Stop(true)`:

After 10 seconds, server side receives log listener closed, all exited old client connections will be stopped by force,
new client connection will be refused.
