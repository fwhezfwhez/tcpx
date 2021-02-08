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
	var certPath = "G:\\go_workspace\\GOPATH\\src\\tcpx\\examples\\modules\\tls\\pem\\cert.pem"
	var keyPath = "G:\\go_workspace\\GOPATH\\src\\tcpx\\examples\\modules\\tls\\pem\\key.pem"

	//var pemPath = "G:\\go_workspace\\GOPATH\\src\\github.com\\fwhezfwhez\\tcpx\\examples\\modules\\tls\\pem"
	if e := r.LoadTLSFile(certPath, keyPath); e != nil {
		panic(e)
	}

	if e := r.ListenAndServe("tcp", ":8080"); e != nil {
		panic(e)
	}
}
