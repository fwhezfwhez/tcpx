<p align="center">
    <a href="github.com/fwhezfwhez/tcpx"><img src="https://user-images.githubusercontent.com/36189053/65203408-cc228800-dabd-11e9-929d-4c9c82b8cdc0.png" width="450"></a>
</p>

<p align="center">
    <a href="https://godoc.org/github.com/fwhezfwhez/tcpx"><img src="http://img.shields.io/badge/godoc-reference-blue.svg?style=flat"></a>
    <a href="https://www.travis-ci.org/fwhezfwhez/tcpx"><img src="https://www.travis-ci.org/fwhezfwhez/tcpx.svg?branch=master"></a>
    <a href="https://gitter.im/fwhezfwhez-tcpx/community"><img src="https://badges.gitter.im/Join%20Chat.svg"></a>
    <a href="https://codecov.io/gh/fwhezfwhez/tcpx"><img src="https://codecov.io/gh/fwhezfwhez/tcpx/branch/master/graph/badge.svg"></a>
</p>

 这是一款go实现的tcp框架，希望可以给大家一个良好的体验

支持以下通信协议
- UDP
- TCP
- KCP(tcpx@v3.0.0 --)

## Declaration

- 由于精力问题，功能的拓展和增强，只会对tcp通信协议来做。UDP只会保留基本的框架预设。
- 由于第三方KCP包版本问题，为了保持对go1.9及以下的可用性，决定在tcpx@v3.0.0之后，放弃对kcp的支持。

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [设计这个框架的缘由?](#%E8%AE%BE%E8%AE%A1%E8%BF%99%E4%B8%AA%E6%A1%86%E6%9E%B6%E7%9A%84%E7%BC%98%E7%94%B1)
- [1. 开始](#1-%E5%BC%80%E5%A7%8B)
    - [必要依赖](#%E5%BF%85%E8%A6%81%E4%BE%9D%E8%B5%96)
    - [压测](#%E5%8E%8B%E6%B5%8B)
- [2. 示例](#2-%E7%A4%BA%E4%BE%8B)
    - [helloworld](#helloworld)
    - [2.1 心跳](#21-%E5%BF%83%E8%B7%B3)
    - [2.2 在线、离线](#22-%E5%9C%A8%E7%BA%BF%E7%A6%BB%E7%BA%BF)
    - [2.3 优雅退出，重启](#23-%E4%BC%98%E9%9B%85%E9%80%80%E5%87%BA%E9%87%8D%E5%90%AF)
    - [2.4 中间件](#24-%E4%B8%AD%E9%97%B4%E4%BB%B6)
    - [2.5 包协议详情](#25-%E5%8C%85%E5%8D%8F%E8%AE%AE%E8%AF%A6%E6%83%85)
    - [2.6 聊天](#26-%E8%81%8A%E5%A4%A9)
    - [2.7 无包协议通讯](#27-%E6%97%A0%E5%8C%85%E5%8D%8F%E8%AE%AE%E9%80%9A%E8%AE%AF)
    - [2.8 用户池](#28-%E7%94%A8%E6%88%B7%E6%B1%A0)
    - [2.9 鉴权](#29-%E9%89%B4%E6%9D%83)
- [3. 使用方法](#3-%E4%BD%BF%E7%94%A8%E6%96%B9%E6%B3%95)
  - [3.1 使用中间件的详情](#31-%E4%BD%BF%E7%94%A8%E4%B8%AD%E9%97%B4%E4%BB%B6%E7%9A%84%E8%AF%A6%E6%83%85)
  - [3.2 省略](#32-%E7%9C%81%E7%95%A5)
  - [3.3 如何准备一段待发送(待回复)的消息?](#33-%E5%A6%82%E4%BD%95%E5%87%86%E5%A4%87%E4%B8%80%E6%AE%B5%E5%BE%85%E5%8F%91%E9%80%81%E5%BE%85%E5%9B%9E%E5%A4%8D%E7%9A%84%E6%B6%88%E6%81%AF)
  - [3.4 选择/自定义包协议的载荷序列化方式](#34-%E9%80%89%E6%8B%A9%E8%87%AA%E5%AE%9A%E4%B9%89%E5%8C%85%E5%8D%8F%E8%AE%AE%E7%9A%84%E8%BD%BD%E8%8D%B7%E5%BA%8F%E5%88%97%E5%8C%96%E6%96%B9%E5%BC%8F)
  - [3.5 非golang如何生成包协议.](#35-%E9%9D%9Egolang%E5%A6%82%E4%BD%95%E7%94%9F%E6%88%90%E5%8C%85%E5%8D%8F%E8%AE%AE)
  - [3.6 不想使用messageID来做路由](#36-%E4%B8%8D%E6%83%B3%E4%BD%BF%E7%94%A8messageid%E6%9D%A5%E5%81%9A%E8%B7%AF%E7%94%B1)
  - [3.7 业务分发?](#37-%E4%B8%9A%E5%8A%A1%E5%88%86%E5%8F%91)
- [4. 使用频率比较高的方法](#4-%E4%BD%BF%E7%94%A8%E9%A2%91%E7%8E%87%E6%AF%94%E8%BE%83%E9%AB%98%E7%9A%84%E6%96%B9%E6%B3%95)
  - [4.1 `tcpx.TcpX`](#41-tcpxtcpx)
  - [4.2 `tcpx.Context`](#42-tcpxcontext)
  - [4.3 `tcpx.Packx`](#43-tcpxpackx)
  - [4.4 `tcpx.Message`](#44-tcpxmessage)
- [5. 协议转换网关](#5-%E5%8D%8F%E8%AE%AE%E8%BD%AC%E6%8D%A2%E7%BD%91%E5%85%B3)
    - [5.1 Gateway pack detail](#51-gateway-pack-detail)
    - [5.2 Gateway unpack detail](#52-gateway-unpack-detail)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## 设计这个框架的缘由?
Golang对tcp的支持十分友好，不过在包的拆组上，官方没有提供明确的方式。所以在对流协议的处理上，需要用特定的包协议来进行完整地拆包和组包。其次，同样一个服务，不同开发人，项目，协议类型的发起服务的写法，很难做到统一，有点群魔乱舞的味道。最后，传统的tcp处理，很容易写成switch方式来分发协议给不同处理函数，这样的处理方式很容易造成多人开发冲突，不美观而且会将项目变得很重。

所以， tcpx提供安全完整的包协议，提供仿http-gin的写法，保持统一，并且，强制使用人按照类似http样式的路由来分发。基于这样的方式来开发，可以让项目变得和http一样简单。

## 1. 开始
`go get github.com/fwhezfwhez/tcpx`

#### 必要依赖
部分本仓库的代码，在运行时，需要安装protoc,protogen-gen,你可以通过下面的链接找到对应的下载方式。

**protoc**: https://github.com/golang/protobuf

**proto-gen-go**:https://github.com/golang/protobuf/tree/master/protoc-gen-go

当你下载并安装成功后，确保以下命令可以得到正确输出:
```
protoc --version
```

#### 压测

https://github.com/fwhezfwhez/tcpx/blob/master/benchmark_test.go

| cases | exec times | cost time per loop | cost mem per loop | cost alloc mem times per loop | url |
|-----------| ---- |------|-------------|-----|-----|
| OnMessage | 2000000 | 643 ns/op | 1368 B/op | 5 allocs/op| [click to location](https://github.com/fwhezfwhez/tcpx/blob/9c70f4bd5a0042932728ed44681ff70d6a22f7e3/benchmark_test.go#L9) |
| Mux without middleware | 2000000 | 761 ns/op | 1368 B/op | 5 allocs/op| [click to location](https://github.com/fwhezfwhez/tcpx/blob/9c70f4bd5a0042932728ed44681ff70d6a22f7e3/benchmark_test.go#L17) |
| Mux with middleware | 2000000 | 768 ns/op | 1368 B/op | 5 allocs/op| [click to location](https://github.com/fwhezfwhez/tcpx/blob/9c70f4bd5a0042932728ed44681ff70d6a22f7e3/benchmark_test.go#L25) |

- cost time per loop: 每次执行的耗时，越小越好
- cost mem per loop: 每次执行的内存占用，越小越好
- cost alloc mem times per loop: 每次执行时申请内存的次数,越小越好

## 2. 示例
https://github.com/fwhezfwhez/tcpx/tree/master/examples/sayHello

#### helloworld
server:
```go
package main

import (
	"fmt"

	"github.com/fwhezfwhez/tcpx"
)

func main() {
	srv := tcpx.NewTcpX(nil)
	srv.OnMessage = func(c *tcpx.Context) {
		var message []byte
		c.Bind(&message)
		fmt.Println(string(message))
	}
	srv.ListenAndServe("tcp", "localhost:8080")
}

```

client:
```go
package main

import (
	"fmt"
	"net"

	"github.com/fwhezfwhez/tcpx"
	//"tcpx"
)

func main() {
	conn, e := net.Dial("tcp", "localhost:8080")

	if e != nil {
		panic(e)
	}
	var message = []byte("hello world")
	buf, e := tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: 1,
		Header:    nil,
		Body:      message,
	}, nil)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	_, e = conn.Write(buf)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
}

```

#### 2.1 心跳
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/heartbeat

- tcpx自带心跳机制，可以通过下面的实例代码开启。在心跳开启时，客户端必须在指定间隔时间内，不断地发送心跳，确保不被服务当作僵尸杀死。
- 心跳机制依赖的是tcpx的mux路由机制。而路由机制是不允许使用Onmessage的，所以使用自带的心跳机制时，需要注意。

**srv side**
```go
    srv := tcpx.NewTcpX(nil)
    srv.HeartBeatModeDetail(true, 10 * time.Second, false, tcpx.DEFAULT_HEARTBEAT_MESSAGEID)
    // srv.OnMessage =nil       Onmessage必须为nil，否则会使得心跳失效
```

**client side**
```go
        var heartBeat []byte
        heartBeat, e = tcpx.PackWithMarshaller(tcpx.Message{
            MessageID: tcpx.DEFAULT_HEARTBEAT_MESSAGEID,
            Header:    nil,
            Body:      nil,
        }, nil)
        for {
            conn.Write(heartBeat)
            time.Sleep(10 * time.Second)
        }
```

**重写心跳逻辑，只要确保在自定义的函数里，执行c.RecvHeartBeat(),过程可以由开发者自我订制**
```go
    srv.RewriteHeartBeatHandler(1300, func(c *tcpx.Context) {
        fmt.Println("rewrite heartbeat handler")
        c.RecvHeartBeat()
    })
```

#### 2.2 在线、离线

- 在线离线需要使用tcpx自带的用户池。

https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/online-offline

#### 2.3 优雅退出，重启
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/graceful

- 优雅退出

在服务被中断时，你可以进行一些收尾工作

- 优雅停止tcp服务(仅仅是tcp的端口暂停请求进入，服务进程不会因此被终结)

- 优雅停止有两种执行策略:

1. `closeAllConnection = false` :停止监听接受新的请求，但是原有的已接入的连接不会受影响，除非它中断了

2. `closeAllConnection = true` : 停止监听新的请求，原有的已接入的连接，会一个个被踢掉.

- 优雅重启:

包含 `graceful stop` and `graceful start` 两个操作

#### 2.4 中间件
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/middleware

这里的例子可以告诉你，如何使用tcpx的中间件

#### 2.5 包协议详情
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/pack-detail

tcpx自带包协议，这里的例子将会描述拆包装包的详情

#### 2.6 聊天
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/chat

这里是使用tcpx，实现了一个简单的聊天

#### 2.7 无包协议通讯
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/raw

如果你不喜欢tcpx自带的包协议，可以利用Raw方式来处理请求。以这种方式接入请求，只有全局中间件(r.UseGlobal)和锚中间件(r.Use)会生效。你会发现，使用srv.OnMessage其实等价于Raw+tcp包协议。

使用该通讯方式时，需要自主读流，拆包解析。

#### 2.8 用户池
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/online-offline

例子和2.2共享，使用offline和online时，需要使用tcpx自带的(很基本功能)的用户连接池。

- 使用自带的池，你需要执行 `srv.WithBuiltInPool(true)`.
- 当服务使用了用户池后，你就可以通过这两个方法来进行离线和在线了 `ctx.Offline()`,`ctx.Online(username string)`.

官方的池不打算做很深的拓展，所以如果池的要求很复杂，还是建议使用者自己实现一个。比如它目前无法做到:

- 一个用户两个渠道同时上线。当然，如果你使用ctx.Online(username:channel)，也可以做到多渠道同时上线

#### 2.9 鉴权
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/auth

鉴权帮助服务端主动隔绝非法连接建立。类似redis, 当连接建立后，必须在指定时间内发送鉴权消息，否则将无法使用服务的任何路由，并且将被杀死。

它和心跳有点不同，因为他一旦收到了正确的鉴权信息，那么它会将该连接视为可信任的，并会结束待鉴权的协程(相比之下，心跳协程会一直持续)。

它和中间件拦截也不同，因为中间件拦截，要求消息体本身，做到像https一样，每一则消息都要带上校验信息才能通过。鉴权只要发出一次正确的请求，则他的生命周期内，都将进行无校验通信(当然，你可以在你的中间件里进行校验通信)。

## 3. 使用方法
用法总结

**使用OnMessage**

- OnMessage 一旦使用，则默认消息内容有开发者自己管理，它将使tcpx的mux路由失效，包括心跳

```go
func main(){
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.OnClose = OnClose
    srv.OnConnect = OnConnect
    srv.OnMessage = OnMessage

    go func(){
        fmt.Println("tcp srv listen on 7171")
        if e := srv.ListenAndServe("tcp", ":7171"); e != nil {
            panic(e)
        }
    }()

    // udp
    go func(){
        fmt.Println("udp srv listen on 7172")
        if e := srv.ListenAndServe("udp", ":7172"); e != nil {
            panic(e)
        }
    }()
    // kcp
    go func(){
        fmt.Println("kcp srv listen on 7173")
        if e := srv.ListenAndServe("kcp", ":7173"); e != nil {
            panic(e)
        }
    }()
    select {}
}

func OnConnect(c *tcpx.Context) {
    fmt.Println(fmt.Sprintf("connecting from remote host %s network %s", c.ClientIP(), c.Network()))
}
func OnClose(c *tcpx.Context) {
    fmt.Println(fmt.Sprintf("connecting from remote host %s network %s has stoped", c.ClientIP(), c.Network())
}
var packx = tcpx.NewPackx(tcpx.JsonMarshaller{})
func OnMessage(c *tcpx.Context) {
    // handle c.Stream
    type ServiceA struct{
        Username string `json:"username"`
    }
    type ServiceB struct{
        ServiceName string `json:"service_name"`
    }

    messageID, e :=packx.MessageIDOf(c.Stream)
    if e!=nil {
        fmt.Println(errorx.Wrap(e).Error())
        return
    }

    switch messageID {
    case 7:
        var serviceA ServiceA
        // block, e := packx.Unpack(c.Stream, &serviceA)
        block, e :=c.Bind(&serviceA)
        fmt.Println(block, e)
        c.Reply(8, "success")
    case 9:
        var serviceB ServiceB
        //block, e := packx.Unpack(c.Stream, &serviceB)
        block, e :=c.Bind(&serviceB)
        fmt.Println(block, e)
        c.JSON(10, "success")
    }
}
```

**使用路由和中间件**

- 下面的例子同时使用了三种中间件，他们是: 全局中间件，锚中间件，路由中间件
- 执行顺序: 全局-->锚--->路由中间件
- 锚中间件可以即启即关，包裹在Use和UnUse之间的路由，将享受到这些中间件的加持

这里介绍一下为什么要将Use()方式使用的中间件成为锚中间件:
我们知道船锚在停靠时落下，他的位置很随意，落下时Use(), 升起时UnUse()，它的行为规律正如同我们的中间件一样，即用即起，所以成为锚中间件。

```go
// 全局中间件
srv.UseGlobal(m1)

// 锚中间件
srv.Use(mkey,m)
srv.UnUse(mkey)

// 路由中间件
srv.AddHandler(m1,m2,m3, handler)
```

```go
func main(){
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.OnClose = OnClose
    srv.OnConnect = OnConnect
    // srv.OnMessage = OnMessage

    srv.UseGlobal(MiddlewareGlobal)
    srv.Use("middleware1", Middleware1, "middleware2", Middleware2)
    srv.AddHandler(1, SayHello)

    srv.UnUse("middleware2")
    srv.AddHandler(3, SayGoodBye)

    if e := srv.ListenAndServe("tcp", ":7171"); e != nil {
        panic(e)
    }
}

func OnConnect(c *tcpx.Context) {
    fmt.Println(fmt.Sprintf("connecting from remote host %s network %s", c.ClientIP(), c.Network()))
}
func OnClose(c *tcpx.Context) {
    fmt.Println(fmt.Sprintf("connecting from remote host %s network %s has stoped", c.ClientIP(), c.Network())
}
// func OnMessage(c *tcpx.Context) {
    // handle c.Stream
// }
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

func SayGoodBye(c *tcpx.Context) {
    var messageFromClient string
    var messageInfo tcpx.Message
    messageInfo, e := c.Bind(&messageFromClient)
    if e != nil {
        panic(e)
    }
    fmt.Println("receive messageID:", messageInfo.MessageID)
    fmt.Println("receive header:", messageInfo.Header)
    fmt.Println("receive body:", messageInfo.Body)

    var responseMessageID int32 = 4
    e = c.Reply(responseMessageID, "bye")
    fmt.Println("reply:", "bye")
    if e != nil {
        fmt.Println(e.Error())
    }
}
func Middleware1(c *tcpx.Context) {
    fmt.Println("I am middleware 1 exampled by 'srv.Use(\"middleware1\", Middleware1)'")
}

func Middleware2(c *tcpx.Context) {
    fmt.Println("I am middleware 2 exampled by 'srv.Use(\"middleware2\", Middleware2),srv.UnUse(\"middleware2\")'")
}

func Middleware3(c *tcpx.Context) {
    fmt.Println("I am middleware 3 exampled by 'srv.AddHandler(5, Middleware3, SayName)'")
}

func MiddlewareGlobal(c *tcpx.Context) {
    fmt.Println("I am global middleware exampled by 'srv.UseGlobal(MiddlewareGlobal)'")
}
```

### 3.1 使用中间件的详情

`全局中间件`:
```go
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.UseGlobal(MiddlewareGlobal)
```
`路由中间件`:
```go
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.AddHandler(5, Middleware3, SayName)
```
`锚中间件`:
```go
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
    srv.Use("middleware1", Middleware1, "middleware2", Middleware2)
    srv.AddHandler(5, SayName)
```


`中间件实现的例子`:
- 使用c.Next()执行下一个装载的中间件，或者路由
- 使用c.Abort()终止链式处理，用法和gin完全一致。

```go
func Middleware1(c *tcpx.Context) {
    fmt.Println("I am middleware 1 exampled by 'srv.Use(\"middleware1\", Middleware1)'")
    // c.Next()
    // c.Abort()
}
```
`执行顺序`:
`全局` -> `锚` -> `路由中间件`.

如果任一中间件的实现里，调用了`c.Abort`，则整条链都会被断掉(当然Abort以前的链还是照常运行的)。
if one of middleware has called `c.Abort()`, middleware chain stops.

**注意**: 再次重申一遍，路由中间件将在启用OnMessage时失效，而全局和锚中间件，横跨所有处理方式，所有协议(包括udp，kcp)，都将生效。

### 3.2 省略

### 3.3 如何准备一段待发送(待回复)的消息?
- tcpx自带包协议，你只需要传入(最少)messageID,消息载荷
- 消息载荷将在序列化时，使用srv声明的序列化方法，默认是json，包括但不限于(json,protobuf, toml, yaml,xml)，并且支持自定义协议
- 自定义协议仅仅是对载荷的序列化可以自设计，不能改变协议规则中 header与 messageid的序列方式。(因为只有json才可以序列化map，而header是map)

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

### 3.4 选择/自定义包协议的载荷序列化方式
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
自定义:
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

### 3.5 非golang如何生成包协议.
tcpx官方提供了go版本的包协议生成，并且提供了python版的参考样例。其他语言，可能需要开发者自己去适配了，当然，tcpx提供http协议网关，你可以参考#5来使用它,

以下是包协议细节
```text
[4]byte -- length             固定4字节，大端编码，指代消息长度(length = packLength-4)
[4]byte -- messageID          固定4字节，大端编码
[4]byte -- headerLength       固定4字节，大端编码
[4]byte -- bodyLength         固定4字节，大端编码
[]byte -- header              json序列化的map结构
[]byte -- body                任意序列化的载荷
```
因为不是每个序列化方式，都支持序列化map结构，所以header固定用json来序列化。


### 3.6 不想使用messageID来做路由
messageID的路由虽然很小巧精致，但是不得不说，他在团队开发时，有一个问题，messageID不能重复，所以开发时，多人需要注意不能混用，重用。messageID的同步会增加心理负担。
通过可读化的文本来路由的方式，已经提上日程，不久后将支持同时使用messageid和文本写路由。有兴趣的读者，可以在这里前瞻:

github.com/fwhezfwhez/wsx

使用时，类似:
```go
// 以下方式规划中
srv.Any("/user/create-user/", mux, handler)
srv.Any("/send-heartbeat/")
```


### 3.7 业务分发?
tcpx目前基于messageid进行路由分发(基于文本分发的方式规划中)
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

## 4. 使用频率比较高的方法
你可以在这里获得所有api: https://godoc.org/github.com/fwhezfwhez/tcpx

这里，将对核心api进行单独讲解。

**参数字段将被省略**
### 4.1 `tcpx.TcpX`
```go
srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
```
| methods | desc |
|--------|--------|
| srv.GlobalUse() | 使用全局中间件|
| srv.Use()| 使用锚中间件 |
| srv.UnUse()| 锚中间件升起。组合之间的路由将执行中间件的逻辑。升起后的锚可以在后续的路由中不断使用和升起。|
| srv.AddHandler()| 增加路由 |
| srv.ListenAndServe()| 监听启动,成功时会阻塞 |

### 4.2 `tcpx.Context`
```go
var c *tcpx.Context
```
| methods | desc |
|---|---|
| c.Bind()| 将数据源绑定进指定结构体 |
| c.Reply() | 回复消息 |
| c.Next() | 中间件放行 |
| c.Abort() | 中间件阻断 |
| c.JSON()| 回复json消息 |
| c.XML()| 回复xml消息 |
| c.YAML()| 回复yaml消息 |
| c.Protobuf()| 回复protobuf消息 |
| c.TOML()| 回复toml消息 |

### 4.3 `tcpx.Packx`
```go
var packx *tcpx.Packx
```
| methods | desc |
|---|---|
| packx.Pack() | 生成包协议 |
| packx.UnPack() | 解析包协议 |
| packx. MessageIDOf()| 从stream从获取它的messageID，要求stream已经按照协议完美切割成一块 |
| packx.LengthOf() | stream块长度-4 |

### 4.4 `tcpx.Message`
```go
var message tcpx.Message
```
| methods | desc |
|---|---|
| message.Get()| 从消息的header中获取一个key对应的值 |
| message.Set() | 将键值设进消息的header|

## 5. 协议转换网关
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
