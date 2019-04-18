package main

import (
	"fmt"
	"tcpx"
)

func main() {
	srv := tcpx.NewTcpX(tcpx.JsonMarshaller{})
	srv.OnClose = OnClose
	srv.OnConnect = OnConnect
	srv.AddHandler(1, SayHello)

	fmt.Println("srv listen on 7171")
	if e:=srv.ListenAndServe("tcp", ":7171");e!=nil{
		panic(e)
	}
}

func OnConnect(c *tcpx.Context){
	fmt.Println(fmt.Sprintf("connecting from remote host %s network %s", c.Conn.RemoteAddr().String(), c.Conn.RemoteAddr().Network()))
}
func OnClose(c *tcpx.Context) {
	fmt.Println(fmt.Sprintf("connecting from remote host %s network %s has stoped", c.Conn.RemoteAddr().String(), c.Conn.RemoteAddr().Network()))
}

func SayHello(c *tcpx.Context) {
	var messageFromClient string
	var messageInfo tcpx.Message
	fmt.Println(c.Stream)
	messageInfo, e := c.Bind(&messageFromClient)
	if e!=nil {
		panic(e)
	}
	fmt.Println("receive messageID:", messageInfo.MessageID)
	fmt.Println("receive header:", messageInfo.Header)
	fmt.Println("receive body:", messageInfo.Body)

	var responseMessageID int32 = 2
	e = c.Reply(responseMessageID, "hello")
	fmt.Println("reply:", "hello")
	if e!=nil {
		panic(e)
	}
}
