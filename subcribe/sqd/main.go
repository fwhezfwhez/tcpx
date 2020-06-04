package main

import (
	"log"
	"sync"
	"tcpx"
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
	srv := tcpx.NewTcpX(nil)
	srv.WithBuiltInPool(true)
	srv.AddHandler(ONLINE, online)
	srv.AddHandler(SUBCRIBE, Subcribe)
	srv.AddHandler(PUBLISH, Publish)
	srv.ListenAndServeTCP("tcp", ":8080")
}

func online(c *tcpx.Context) {
	type Param struct {
		Username string `json:"username"`
	}

	var param Param
	if _, e := c.Bind(&param); e != nil {
		log.Println(e.Error())
		return
	}

	c.Online(param.Username)

	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second)
	// 		ok := c.IsOnline()
	// 		fmt.Println(fmt.Sprintf("%v", ok))
	// 	}
	// }()

	c.JSON(INFO, "online success")
}

type Channel string
type Username string

var subscribers = make(map[Channel][]Username)
var l sync.RWMutex

func Subcribe(c *tcpx.Context) {
	type Param struct {
		Channel string `json:"channel"`
	}
	var param Param
	if _, e := c.Bind(&param); e != nil {
		c.JSON(ERROR, e.Error())
		return
	}

	username, ok := c.Username()
	if !ok {
		log.Println("not online yet, username not found")
		c.JSON(ERROR, "not online yet, username not found")
		return
	}
	l.Lock()
	users, ok := subscribers[Channel(param.Channel)]
	if !ok {
		subscribers[Channel(param.Channel)] = []Username{Username(username)}
	} else {
		users = append(users, Username(username))
		subscribers[Channel(param.Channel)] = users
	}
	l.Unlock()

	c.JSON(INFO, "subcribe success")
}

func Publish(c *tcpx.Context) {
	type Param struct {
		Channel string `json:"channel"`
		Message []byte `json:"message"`
	}
	var param Param
	if _, e := c.Bind(&param); e != nil {
		log.Println(e.Error())
		return
	}

	l.RLock()
	users, ok := subscribers[Channel(param.Channel)]
	l.RUnlock()

	if !ok {
		c.JSON(INFO, "no one has subscribe the topic, do nothing")
		return
	}

	for i, _ := range users {
		go c.SendToUsername(string(users[i]), RECV_SUBCRIBE, param.Message)
	}
}
