package main

import (
	"fmt"
	"net"
	"tcpx"
	"time"

	"github.com/fwhezfwhez/errorx"
)

const (
	INFO  = 0
	ERROR = -1

	ONLINE   = 1
	SUBCRIBE = 3

	RECV_SUBCRIBE = 4

	PUBLISH = 5
)

func main() {
	conn, e := tcpx.TCPConnect("tcp", "localhost:8080")
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	go recv(conn)

	if e := tcpx.PipeJSON(conn,

		// online
		ONLINE,
		struct {
			Username string `json:"username"`
		}{Username: "client1"},

		// subscribe
		SUBCRIBE,
		struct {
			Channel string `json:"channel"`
		}{Channel: "weather-report"},
	); e != nil {
		fmt.Println(e.Error())
		return
	}

	time.Sleep(20 * time.Second)

	if e := tcpx.PipeJSON(conn,

		// online
		ONLINE,
		struct {
			Username string `json:"username"`
		}{Username: "client1"},

		// subscribe
		SUBCRIBE,
		struct {
			Channel string `json:"channel"`
		}{Channel: "weather-report"},
	); e != nil {
		fmt.Println(e.Error())
		return
	}

	select {}
}

func recv(conn net.Conn) {
	for {
		block, e := tcpx.FirstBlockOf(conn)
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		go handle(block)
	}
}

func handle(block []byte) {
	msgID, e := tcpx.MessageIDOf(block)
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}
	bodyBuf, e := tcpx.BodyBytesOf(block)
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}

	switch msgID {
	case INFO:
		var message string
		if e := tcpx.BindJSON(bodyBuf, &message); e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		fmt.Println("[INFO]", message)
	case ERROR:
		var message string
		if e := tcpx.BindJSON(bodyBuf, &message); e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		fmt.Println("[ERROR]", message)
	case RECV_SUBCRIBE:
		var message []byte
		if e := tcpx.BindJSON(bodyBuf, &message); e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		fmt.Println("[RECV_CHANNEL_MESSAGE]", string(message))
	}
}
