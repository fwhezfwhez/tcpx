package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/fwhezfwhez/tcpx"
	//"tcpx"
	"log"
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
		if e := srv.Stop(false); e != nil {
			log.Println(fmt.Sprintf("%s \n %s", e.Error(), debug.Stack()))
			return
		}
		// operate between stop and start
		// do something
		fmt.Println("before start, print ok")

		// after 10 seconds start again
		time.Sleep(10 * time.Second)
		if e := srv.Start(); e != nil {
			log.Println(fmt.Sprintf("%s \n %s", e.Error(), debug.Stack()))
			return
		}

		// or call `Restart()` equals to above `Close` and `Start`
		//if e := srv.Restart(false, func() {
		//	fmt.Println("before start, print ok")
		//}); e != nil {
		//	fmt.Println(e.Error())
		//}
	}()

	select {}
}
