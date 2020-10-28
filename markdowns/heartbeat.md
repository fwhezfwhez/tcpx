#### Heartbeat

https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/heartbeat

tcpx has built-in heartbeat handler. Default heartbeat messageID is 1392.It means client should send heartbeat pack in specific interval.When fail received more than 3 times, connection will break by server.

**srv side**
```go
    srv := tcpx.NewTcpX(nil)
    srv.HeartBeatModeDetail(true, 10 * time.Second, false, tcpx.DEFAULT_HEARTBEAT_MESSAGEID)
```

**client side**
```go
        var heartBeat []byte
        heartBeat, e = tcpx.PackWithMarshaller(tcpx.Message{
            MessageID: tcpx.DEFAULT_HEARTBEAT_MESSAGEID,
            Header:    nil,
            Body:      nil,
        }, nil)
        for {
            conn.Write(heartBeat)
            time.Sleep(10 * time.Second)
        }
```

**rewrite heartbeat handler**
```go
    srv.RewriteHeartBeatHandler(1300, func(c *tcpx.Context) {
        fmt.Println("rewrite heartbeat handler")
        c.RecvHeartBeat()
    })
```