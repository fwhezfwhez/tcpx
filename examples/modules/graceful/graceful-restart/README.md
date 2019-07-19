## graceful-restart

#### step
`go run main.go`

`cd client`

`go run client.go` **accept**

after 10s
`go run client.go` **refused**

after another 10s
`go run client.go` **accept**

#### result
three client connections:

first connection is accepted.

second connection is refused because server has been graceful stopped.

third connection is accepted because server start again.
