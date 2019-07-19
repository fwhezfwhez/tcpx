package main

import (
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"sync/atomic"
	//"tcpx"
)

var requestTimes int32

func main() {
	srv := tcpx.NewTcpX(nil)
	srv.UseGlobal(countRequestTime)
	srv.AddHandler(3, getRequestTime)

	fmt.Println("tcp listen on :8080")
	if e := srv.ListenAndServe("tcp", ":8080"); e != nil {
		panic(e)
	}
}

func countRequestTime(c *tcpx.Context) {
	atomic.AddInt32(&requestTimes, 1)
}
func getRequestTime(c *tcpx.Context) {
	c.JSON(4, tcpx.H{"message": "success", "request_times": requestTimes})
}
