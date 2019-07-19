package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	//"tcpx"
	"time"
)

func main() {
	srv := tcpx.NewTcpX(nil)
    // srv.WithBroadCastSignal(true)
    srv.WithBuiltInPool(true)

	srv.OnConnect = func(c *tcpx.Context) {
		c.Online("hehe")
	}
	// start server
	go func() {
		fmt.Println("tcp listen on :8080")
		srv.ListenAndServe("tcp", ":8080")
	}()

	// after 10 seconds and stop it
    go func() {
        time.Sleep(10 * time.Second)
        if e:=srv.Stop(false); e!=nil {
        	fmt.Println(errorx.Wrap(e).Error())
        	return
		}
		//
		//if e:=srv.Stop(true); e!=nil {
		//	fmt.Println(errorx.Wrap(e).Error())
		//	return
		//}
	}()

	select{}
}
