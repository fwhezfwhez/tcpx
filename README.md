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


### How client (not only golang) builds expected stream?
Tcpx now only provide `packx` realized in golang to build a client sender.If you wants to send message from other language client, you'll have two ways:
1. Be aware of messageID block system and build expected stream in specific language.Refer -> [2.5 pack-detail](#25-pack-detail).
2. Using http gateway,refers to **[5. cross-language gateway](#5-cross-language-gateway)**

`messageID block system`:
```text
[4]byte -- length             fixed_size,binary big endian encode
[4]byte -- messageID          fixed_size,binary big endian encode
[4]byte -- headerLength       fixed_size,binary big endian encode
[4]byte -- bodyLength         fixed_size,binary big endian encode
[]byte -- header              marshal by json
[]byte -- body                marshal by marshaller
```
Since not all marshal ways support marshal map, header are fixedly using json.
Here are some language building stream:
java:
```java
//
```
js:
```js
//
```
ruby:
```
//
```
Welcome to provides all language pack example via pull request, you can valid you result stream refers to unpack http gateway **[5. cross-language gateway](#5-cross-language-gateway)**，

## Frequently used methods
All methods can be refered in https://godoc.org/github.com/fwhezfwhez/tcpx
Here are those frequently used methods apart by their receiver type.
**args omit**
### `tcpx.TcpX`
```go
srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
```
| methods | desc |
|--------|--------|
| srv.GlobalUse() | use global middleware|
| srv.Use()| use a middleware|
| srv.UnUse()| unUse a middleware, handlers added before this still work on unUsed middleware, handlers after don't|
| srv.AddHandler()| add routed handler by messageID(int32) |
| srv.ListenAndServe()| start listen on |
### `tcpx.Context`
```go
var c *tcpx.Context
```
| methods | desc |
|---|---|
| c.Bind()| bind data of stream into official message type |
| c.Reply() | reply to client via c.Conn, marshalled by c.Packx.Marshaller |
| c.Next() | middleware goes to next |
| c.Abort() | middleware chain stops|
| c.JSON()| reply to client via c.Conn, marshalled by tcpx.JsonMarshaller |
| c.XML()| reply to client via c.Conn, marshalled by tcpx.XmlMarshaller |
| c.YAML()| reply to client via c.Conn, marshalled by tcpx.YamlMarshaller |
| c.Protobuf()| reply to client via c.Conn, marshalled by tcpx.ProtobufMarshaller |
| c.TOML()| reply to client via c.Conn, marshalled by tcpx.TomlMarshaller |

### `tcpx.Packx`
```go
var packx *tcpx.Packx
```
| methods | desc |
|---|---|
| packx.Pack() | pack data into expected stream |
| packx.UnPack() | reverse above returns official message type |
| packx. MessageIDOf()| get messageID of a stream block |
| packx.LengthOf() | length of stream except total length, total length +4 or len(c.Stream) |

### `tcpx.Message`
```go
var message tcpx.Message
```
| methods | desc |
|---|---|
| message.Get()| get header value by key |
| message.Set() | set header value |

## Cross-language gateway
gateway repo:
https://github.com/fwhezfwhez/tcpx/tree/master/gateway/pack-transfer

example:
https://github.com/fwhezfwhez/tcpx/tree/master/examples/use-gateway

`go run main.go -port 7000`  run the gateway locally in port 7000 or else.

#### 5.1 Gateway pack detail
**note: Each message should call once**
```url
POST http://localhost:7000/gateway/pack/transfer/
application/json
```
body:
```json
{
    "marshal_name":<marshal_name>,
    "stream": <stream>,
    "message_id": <message_id>,
    "header": <header>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| marshal_name | string |ranges in `"json","xml", "toml", "yaml", "protobuf"`| "json"|yes|
| stream | []byte | stream should be well marshalled by one of marshal_name | | yes|
|message_id | int32 | int32 type messageID| 1 | yes|
| header | map/object | key-value pairs | {"k1":"v1"}| no|

returns:
```json
{
    "message":<message>,
    "stream":<stream>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| message | string |"success" when status 200, "success", "error message" when 400/500 | "success"|yes|
| stream | []byte | packed stream,when error or status not 200, no stream field | | no|

example:

payload:
```go
{"username": "hello, tcpx"}   ---json-->  "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0="
```

request:
```json
{
    "marshal_name": "json",
    "stream": "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0=",
    "message_id": 1,
    "header": {
      "api": "/pack/"
    }
}
```

example response:
```json
{
    "stream": "AAAANgAAAAEAAAAQAAAAGnsiYXBpIjoiL3BhY2svIn17InVzZXJuYW1lIjoiaGVsbG8sIHRjcHgifQ=="
}
```
#### 5.2 Gateway unpack detail
**note: able to unpack many messages once.**
```url
POST http://localhost:7000/gateway/unpack/transfer/
application/json
```
body:
```json
{
    "marshal_name": <marshal_name>,
    "stream": <stream>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| marshal_name | string |ranges in `"json","xml", "toml", "yaml", "protobuf"`| "json"|yes|
| stream | []byte | packed stream| | no|

returns:
```json
{
    "message": <message>,
    "blocks" <blocks>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| message | string |"success" when status 200, "success", "error message" when 400/500 | "success"|yes|
| blocks | []block | unpacked blocks, when status not 200, no this field| | no|
|block| obj | each message block information, when status not 200,no this field | ++ look below++ | no|

block example:
```json
{
    "message_id": 1,
    "header": {"k1":"v1"},
    "marshal_name": "json",
    "stream": "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0="
}
```
example request:
```json
{
    "marshal_name": "json",
    "stream": "AAAANgAAAAEAAAAQAAAAGnsiYXBpIjoiL3BhY2svIn17InVzZXJuYW1lIjoiaGVsbG8sIHRjcHgifQ=="
}
```
example response:
```json
{
    "message": "success",
    "blocks": [
      {
        "message_id": 1,
        "header": {
          "k1": "v1"
        },
        "marshal_name": "json",
        "stream": "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0="
      }
    ]
}
```

to payload:
```go
"eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0="   ---json-->  {"username": "hello, tcpx"}
```
