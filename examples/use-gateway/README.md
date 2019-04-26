use-gateway shows how clients using http gateway to build expected stream,supporting json,xml,toml,yaml,protobuf marshal types.

`cd ${GOPATH}/srv/tcpx`

`cd gateway/pack-transfer`

`go run main.go` start gateway program default 7000 port

`cd ${GOPATH}/srv/tcpx`

`cd examples/use-gateway`

`go run client.go` connect to port 7000


