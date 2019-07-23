package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	conn, e := net.Dial("tcp", "localhost:6631")
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	go Recv(conn)
	conn.Write([]byte("hello,I am client."))

	select {}
}

func Recv(conn net.Conn) {
	for {
		var buf = make([]byte,500)
		n,e:=conn.Read(buf)
		if e!=nil {
			fmt.Println(e.Error())
			os.Exit(1)
		}
		fmt.Println(string(buf[:n]))
	}
}
