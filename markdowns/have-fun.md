## Ok, Let's have fun!

## 1. Shortcut
#### 1.1 start a tcp server
- tcp listen on 7071 port

```go
srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
...
srv.ListenAndServe("tcp", ":7071")
```

#### 1.2 Add handler
- login handler handles message whose messageID is 3.
- genToken handler handles message whose urlPattern is '/get-token/'
```go
srv.AddHandler(3, login)
srv.Any("/get-token/", genToken)
```

#### 1.3 A middleware
- global middleware logIP. Take effect for all handlers.
- anchor middleware limitRate. Take effect for handlers wrapped by Use and UnUse, or just behind Use without UnUse.
- self middleware limitTimes. Take effect only for specific handler.
```go
srv.UseGlobal(logIP)       // global middleware
srv.Use("limit_rate_middleware",limitRate)
srv.AddHandler(4, limitTimes, increaseMoney)    //
srv.UnUse("limit_rate_middleware")
```
#### 1.4 A handler
```go
func LogIP(c *tcpx.Context) {
    fmt.Printf("recv request from ip: %s\n", c.ClientIP())
}
```

#### 1.5 Bind model
```go
func Login(c *tcpx.Context) {
    type LoginRequest struct{
         Username string `json:"username`
    }

    var lr LoginRequest
    c.Bind(&lr)
}
```

#### 1.6 Reply
- c.JSON(), reply a messageID type message, marshalled by JSON
- c.JSONURLPattern(), replay a urlPattern type message, marshalled by JSON, response's urlPattern will share with request's urlPattern
```go
func Login(c *tcpx.Context) {
    c.JSON(4, map[string]interface{}{
        "token":"abc"
    })

    c.JSONURLPattern(map[string]interface{}{
        "token":"abc"
    })
}
```
