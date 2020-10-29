<p align="center">
    <a href="github.com/fwhezfwhez/tcpx"><img src="https://user-images.githubusercontent.com/36189053/65203408-cc228800-dabd-11e9-929d-4c9c82b8cdc0.png" width="450"></a>
</p>

<p align="center">
    <a href="https://godoc.org/github.com/fwhezfwhez/tcpx"><img src="http://img.shields.io/badge/godoc-reference-blue.svg?style=flat"></a>
    <a href="https://www.travis-ci.org/fwhezfwhez/tcpx"><img src="https://www.travis-ci.org/fwhezfwhez/tcpx.svg?branch=master"></a>
    <a href="https://gitter.im/fwhezfwhez-tcpx/community"><img src="https://badges.gitter.im/Join%20Chat.svg"></a>
    <a href="https://codecov.io/gh/fwhezfwhez/tcpx"><img src="https://codecov.io/gh/fwhezfwhez/tcpx/branch/master/graph/badge.svg"></a>
</p>

A very convenient tcp framework in golang.

- [Have fun](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/have-fun.md)
- [Heartbeat](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/heartbeat.md)
- [Auth](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/auth.md)
- [User pool](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/user-pool.md)
- [Graceful](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/graceful.md)
- [Middleware](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/middleware.md)
- [Message](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/message.md)
- [Marshaller](https://github.com/fwhezfwhez/tcpx/tree/master/markdowns/marshaller.md)

## Start
`go get github.com/fwhezfwhez/tcpx`

#### Dependency
if you want to run program in this repo,you should prepare protoc,proto-gen-go environment.
It's good to compile yourself from these repos,but there is already release versions referring to their doc.
Make sure run `protoc --version` available.

**protoc**: https://github.com/golang/protobuf

**proto-gen-go**:https://github.com/golang/protobuf/tree/master/protoc-gen-go

#### Benchmark

https://github.com/fwhezfwhez/tcpx/blob/master/benchmark_test.go

| cases | exec times | cost time per loop | cost mem per loop | cost object num per loop | url |
|-----------| ---- |------|-------------|-----|-----|
| OnMessage | 2000000 | 643 ns/op | 1368 B/op | 5 allocs/op| [click to location](https://github.com/fwhezfwhez/tcpx/blob/9c70f4bd5a0042932728ed44681ff70d6a22f7e3/benchmark_test.go#L9) |
| Mux without middleware | 2000000 | 761 ns/op | 1368 B/op | 5 allocs/op| [click to location](https://github.com/fwhezfwhez/tcpx/blob/9c70f4bd5a0042932728ed44681ff70d6a22f7e3/benchmark_test.go#L17) |
| Mux with middleware | 2000000 | 768 ns/op | 1368 B/op | 5 allocs/op| [click to location](https://github.com/fwhezfwhez/tcpx/blob/9c70f4bd5a0042932728ed44681ff70d6a22f7e3/benchmark_test.go#L25) |

#### Pack
Tcpx has its well-designed pack. To focus on detail, you can refer to:
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/pack-detail

```text
[4]byte -- length             fixed_size,binary big endian encode
[4]byte -- messageID          fixed_size,binary big endian encode
[4]byte -- headerLength       fixed_size,binary big endian encode
[4]byte -- bodyLength         fixed_size,binary big endian encode
[]byte -- header              marshal by json
[]byte -- body                marshal by marshaller
```

#### Chat
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/chat

It examples a chat using tcpx.

#### Raw
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/raw

It examples how to send stream without rule, nothing to do with `messageID system`. You can send all stream you want. Global middleware and anchor middleware are still working as the example said.

`messageID block system`:
```text
[4]byte -- length             fixed_size,binary big endian encode
[4]byte -- messageID          fixed_size,binary big endian encode
[4]byte -- headerLength       fixed_size,binary big endian encode
[4]byte -- bodyLength         fixed_size,binary big endian encode
[]byte -- header              marshal by json
[]byte -- body                marshal by marshaller
```