#### User Pool
https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/online-offline

Tcpx has its built-in pool to help manage online and offline users. Note that :

- To use built-in pool, you need to run `srv.WithBuiltInPool(true)`.
- To online/offline a user, you can do it like `ctx.Offline()`,`ctx.Online(username string)`.

Official built-in pool will not extend much. If it doesn't fit your requirement, you should design your own pool.

```go
func main() {
	srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})

	srv.WithBuiltInPool(true)

	// srv.AddHandler(1, online)
	srv.Any("/login/", online)
	srv.ListenAndServe("tcp", ":8102")
}

func online(c *tcpx.Context) {
	type Login struct {
		Username string `json:"username"`
	}
	var login Login
	if _, e := c.Bind(&login); e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}
	c.Online(login.Username)
	fmt.Println("online success")
	// c.Offline()
}

```
