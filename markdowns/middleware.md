## Middleware
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/middleware

### GlobalMiddleware

```go
func main(){
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

    srv.UseGlobal(MiddlewareGlobal)
    srv.AddHandler(1, SayHello)

    srv.ListenAndServe("tcp", ":7171")
}


func SayHello(c *tcpx.Context) {
     fmt.Println("hello")
}

func MiddlewareGlobal(c *tcpx.Context) {
    fmt.Println("I am global middleware exampled by 'srv.UseGlobal(MiddlewareGlobal)'")
}
```

### Anchor Middleware
```go
func main(){
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

    srv.Use("mid", MiddlewareAnchor)
    srv.AddHandler(1, SayHello)
    srv.UnUse("mid")

    srv.AddHandler(4, SayGoodBye)

    srv.ListenAndServe("tcp", ":7171")
}


func SayHello(c *tcpx.Context) {
     fmt.Println("hello")
}

func SayGoodBye(c *tcpx.Context) {
     fmt.Println("bye")
}

func MiddlewareAnchor(c *tcpx.Context) {
    fmt.Println("I am anchor middleware)
}
```

### Self middleware
```go
func main(){
    srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

    srv.AddHandler(1,MiddlewareSelf, SayHello)

    srv.ListenAndServe("tcp", ":7171")
}


func SayHello(c *tcpx.Context) {
     fmt.Println("hello")
}


func MiddlewareSelf(c *tcpx.Context) {
    fmt.Println("I am self middleware)
}
```