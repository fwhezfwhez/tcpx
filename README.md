<p align="center">
    <a href="github.com/fwhezfwhez/tcpx"><img src="http://i1.bvimg.com/684630/1866f4faad40119b.png" width="450"></a>
</p>

<p align="center">
    <a href="https://godoc.org/github.com/fwhezfwhez/tcpx"><img src="http://img.shields.io/badge/godoc-reference-blue.svg?style=flat"></a>
    <a href="https://www.travis-ci.org/fwhezfwhez/tcpx"><img src="https://www.travis-ci.org/fwhezfwhez/tcpx.svg?branch=master"></a>
    <a href="https://gitter.im/fwhezfwhez-tcpx/community"><img src="https://badges.gitter.im/Join%20Chat.svg"></a>
</p>

a very convenient tcp framework in golang.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [tcpx](#tcpx)
  - [Why designing tcp framwork rather than the official?](#why-designing-tcp-framwork-rather-than-the-official)
  - [1. Start](#1-start)
  - [2. Example](#2-example)
  - [3. Ussages](#3-ussages)
    - [3.1 How to add middlewares?](#31-how-to-add-middlewares)
    - [3.2 When to use OnMessage callback?](#32-when-to-use-onmessage-callback)
    - [3.3 How to design a message?](#33-how-to-design-a-message)
    - [3.4 How to specific marshal type?](#34-how-to-specific-marshal-type)
    - [3.5 How client (not only golang) builds expected stream?](#35-how-client-not-only-golang-builds-expected-stream)
    - [3.6 Can user design his own message rule rather than tcpx.Message pack rule?](#36-can-user-design-his-own-message-rule-rather-than-tcpxmessage-pack-rule)
    - [3.7 How to separate handlers?](#37-how-to-separate-handlers)
  - [4. Frequently used methods](#4-frequently-used-methods)
    - [4.1 `tcpx.TcpX`](#41-tcpxtcpx)
    - [4.2 `tcpx.Context`](#42-tcpxcontext)
    - [4.3 `tcpx.Packx`](#43-tcpxpackx)
    - [4.4 `tcpx.Message`](#44-tcpxmessage)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Why designing tcp framwork rather than the official?
Golang has greate support of tcp protocol in official libraries, but users still need to consider details, most profiling way will make project heavier and heavier.Tpcx aims to use tcp in a most friendly way.Most ussage paterns are like `github.com/gin-gonic/gin`.Users don't consider details. All they are advised touching is a context, most apis in `gin` are also accessable in tcpx.

## 1. Start
`go get github.com/fwhezfwhez/tcpx`

## 2. Example
https://github.com/fwhezfwhez/tcpx/tree/master/examples/sayHello

## 3. Ussages
### 3.1 How to add middlewares?
Middlewares in tcpx has three types: `GlobalTypeMiddleware`, `MessageIDSelfRelatedTypeMiddleware`,`AnchorTypeMiddleware`.
`GlobalTypeMiddleware`:
```go
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.UseGlobal(MiddlewareGlobal)
```
`MessageIDSelfRelatedTypeMiddleware`:
```go
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.AddHandler(5, Middleware3, SayName)
```
`AnchorTypeMiddleware`:
```go
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.Use("middleware1", Middleware1, "middleware2", Middleware2)
    srv.AddHandler(5, SayName)
```
`middleware example`:
```go
func Middleware1(c *tcpx.Context) {
    fmt.Println("I am middleware 1 exampled by 'srv.Use(\"middleware1\", Middleware1)'")
    // c.Next()
    // c.Abort()
}
```
`middleware order`:
`GlobalTypeMiddleware` -> `AnchorTypeMiddleware` -> `MessageIDSelfRelatedTypeMiddleware`.
if one of middleware has called `c.Abort()`, middleware chain stops.

**ATTENTION**: If `srv.OnMessage` is not nil, only `GlobalTypeMiddleware` and `AnchorTypeMiddleware` will make sense regardless of `AnchorTypeMiddleware` being UnUsed or not.

### 3.2 When to use OnMessage callback?
`OnMessage` 's minimum unit block is **each message**, when`OnMessage` is not nil, `mux` will lose its effects.In the mean time, global middlewares and anchor middlewares will all make sense regardless of anchor middlewares being unUsed or not.
Here is part of source code:
```go
go func(ctx *Context, tcpx *TcpX) {
        if tcpx.OnMessage != nil {
            ...
        } else {
            messageID, e := tcpx.Packx.MessageIDOf(ctx.Stream)
            if e != nil {
                Logger.Println(errorx.Wrap(e).Error())
                return
            }
            handler, ok := tcpx.Mux.Handlers[messageID]
            if !ok {
                Logger.Println(fmt.Sprintf("messageID %d handler not found", messageID))
                return
            }
            ...
        }
    }(ctx, tcpx)
```
As you can see,it's ok if you do it like:
```go
func main(){
    ...
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.OnMessage = onMessage
    ...
}
func onMessage(c *tcpx.Context){
    func(stream []byte){
        // handle raw stream
    }(c.Stream)
}
```
**Attention**: Stream has been packed per request, no pack stuck probelm. 

### 3.3 How to design a message?
You don't need to design message block yourself.Instead do it like:
client
```go
func main(){
    var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})
    buf1, e := packx.Pack(5, "hello,I am client xiao ming")
    buf2, e := packx.Pack(7, struct{
    Username string
    Age int
    }{"xiaoming", 5})
    ...
}
```
If you're not golang client, see **[3.5 How client (not only golang) builds expected stream?](#35-how-client-not-only-golang-builds-expected-stream)**
### 3.4 How to specific marshal type?
Now, tcpx supports json,xml,protobuf,toml,yaml like:
client
```go
var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})
// var packx = tcpx.NewPackx(tcpx.XmlMarshaller{})
// var packx = tcpx.NewPackx(tcpx.ProtobufMarshaller{})
// var packx = tcpx.NewPackx(tcpx.TomlMarshaller{})
// var packx = tcpx.NewPackx(tcpx.YamlMarshaller{})
```
server
```go
srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
// srv := tcpx.NewTcpX(tcpx.XmlMarshaller{})
// srv := tcpx.NewTcpX(tcpx.ProtobufMarshaller{})
// srv := tcpx.NewTcpX(tcpx.TomlMarshaller{})
// srv := tcpx.NewTcpX(tcpx.YamlMarshaller{})
```
if you want any marshal way else, design it like:
```go
type OtherMarshaller struct{}
func (om OtherMarshaller) Marshal(v interface{}) ([]byte, error) {
    return []byte(""), nil
}
func (om OtherMarshaller) Unmarshal(data []byte, dest interface{}) error {
    return nil
}
func (om OtherMarshaller) MarshalName() string{
    return "other_marshaller"
}
```
client
```go
var packx = tcpx.NewPackx(OtherMarshaller{})
```
server
```go
srv := tcpx.NewTcpX(tcpx.OtherMarshaller{})
```

### 3.5 How client (not only golang) builds expected stream?
Tcpx now only provide `packx` realized in golang to build a client sender.If you wants to send message from other language client, you should be aware of messageID block system.

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
Welcome to provides all language pack example via pull request, you can valid you result stream via [updating...](https://www.baidu.com)

### 3.6 Can user design his own message rule rather than tcpx.Message pack rule?
It's on developing. future tcpx will support two other pack rule:
1. Based on messageID block system, users design inner protocols in mesesage.Body.
2. Completely different from messageID block system, users redesign stream rule.

messageID block system can refer to **[3.5 How client (not only golang) builds expected stream?](#35-how-client-not-only-golang-builds-expected-stream)**.

### 3.7 How to separate handlers?
tcpx's official advised routing way is separating handlers by messageID, like
```go
func main(){
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    // request messageID 1
    // response messageID 2
    srv.AddHandler(1, SayHello)
    if e := srv.ListenAndServe("tcp", ":7171"); e != nil {
        panic(e)
    }
}
func SayHello(c *tcpx.Context) {
    var messageFromClient string
    var messageInfo tcpx.Message
    messageInfo, e := c.Bind(&messageFromClient)
    if e != nil {
        panic(e)
    }
    fmt.Println("receive messageID:", messageInfo.MessageID)
    fmt.Println("receive header:", messageInfo.Header)
    fmt.Println("receive body:", messageInfo.Body)

    var responseMessageID int32 = 2
    e = c.Reply(responseMessageID, "hello")
    fmt.Println("reply:", "hello")
    if e != nil {
        fmt.Println(e.Error())
    }
}

```

## 4. Frequently used methods
All methods can be refered in https://godoc.org/github.com/fwhezfwhez/tcpx
Here are those frequently used methods apart by their receiver type.
**args omit**
### 4.1 `tcpx.TcpX`
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
### 4.2 `tcpx.Context`
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

### 4.3 `tcpx.Packx`
```go
var packx *tcpx.Packx
```
| methods | desc |
|---|---|
| packx.Pack() | pack data into expected stream |
| packx.UnPack() | reverse above returns official message type |
| packx. MessageIDOf()| get messageID of a stream block |
| packx.LengthOf() | length of stream except total length, total length +4 or len(c.Stream) |

### 4.4 `tcpx.Message`
```go
var message tcpx.Message
```
| methods | desc |
|---|---|
| message.Get()| get header value by key |
| message.Set() | set header value |
