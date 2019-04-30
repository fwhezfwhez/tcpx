package tcpx

import (
	"fmt"
	"testing"
	"time"
)

func TestTcpX_ListenAndServe(t *testing.T) {
	var onConnect = func(c *Context)  {
		fmt.Println(fmt.Sprintf("connecting from remote host %s network %s", c.Conn.RemoteAddr().String(), c.Conn.RemoteAddr().Network()))

	}
	var onClose = func(c *Context)  {
		fmt.Println(fmt.Sprintf("connecting from remote host %s network %s has stoped", c.Conn.RemoteAddr().String(), c.Conn.RemoteAddr().Network()))

	}
	var sayHello_1 = func(c *Context) {
		fmt.Println("hello")

	}
	var sayGoodBye_2 = func(c *Context) {
		fmt.Println("good bye")

	}
	tcpx := NewTcpX(JsonMarshaller{})
	tcpx.OnConnect = onConnect
	tcpx.OnClose = onClose
	tcpx.AddHandler(1, sayHello_1)
	tcpx.AddHandler(2, sayGoodBye_2)

	fmt.Println("开始监听: tcp 7676")
	go func() {
		e := tcpx.ListenAndServe("tcp", ":7676")
		if e != nil {
			fmt.Println(e.Error())
			return
		}
	}()
	time.Sleep(10 * time.Second)
}
func TestTcpX_Clone(t *testing.T) {
	var former = NewTcpX(JsonMarshaller{})

	fmt.Println(former.Clone())
}
