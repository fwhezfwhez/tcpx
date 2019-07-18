package main

import (
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	//"tcpx"
)

func main() {
	srv := tcpx.NewTcpX(nil)
	srv.BeforeExit(
		func() {
			fmt.Println("clear online cache")
		},
		func() {
			fmt.Println("job2 done")
		},
	)
	srv.ListenAndServe("tcp", ":8080")
}
