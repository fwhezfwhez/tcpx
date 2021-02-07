## tls
Tls help you to enable tls for tcp server. To well using tls, you should know:

- If your ISP provides domain mounted tls certification to your host, and all of your app clients connect to your server via this domain, then no need to use tcpx to start a tls server.
- If you don't use certification provided by ISP(You got from an authorized open-source-free third-party orgnization with `key.pem` and `cert.pem`), then you can use tcpx to run with it:
```go
package main

import (
	"fmt"
	"github.com/fwhezfwhez/tcpx"
)

func main() {
	r := tcpx.NewTcpX(nil)
	r.Any("/login/", func(c *tcpx.Context) {
		fmt.Printf("recv a login")
	})

	// TODO You might modify here to locate your pem files' real path

    //var certPath = "G:\\go_workspace\\GOPATH\\src\\github.com\\fwhezfwhez\\tcpx\\examples\\modules\\tls\\pem\\cert.pem"
	var certPath = "G:\\go_workspace\\GOPATH\\src\\tcpx\\examples\\modules\\tls\\pem\\cert.pem"

    //var keyPath = "G:\\go_workspace\\GOPATH\\src\\github.com\\fwhezfwhez\\tcpx\\examples\\modules\\tls\\pem\\key.pem"
	var keyPath = "G:\\go_workspace\\GOPATH\\src\\tcpx\\examples\\modules\\tls\\pem\\key.pem"

	//var pemPath = "G:\\go_workspace\\GOPATH\\src\\github.com\\fwhezfwhez\\tcpx\\examples\\modules\\tls\\pem"
	if e := r.LoadTLSFile(certPath, keyPath); e != nil {
		panic(e)
	}

	if e := r.ListenAndServe("tcp", ":8080"); e != nil {
		panic(e)
	}
}

```

Example:
[tls example](https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/tls)
