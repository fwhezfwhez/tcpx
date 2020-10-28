## message
tcpx message has 2 types, messageID-type and url-type.

messageID-type message(assume messageID=32) will be routed by `srv.AddHandler(32, login)`.

url-type message(assume urlPattern="/login") will be routed by `srv.Any("/login", login)`

#### go client
```go
message1 := tcpx.NewMessage(12, map[string]interface{}{"username":"tommy"})
buf1, _ := messag1.Pack(tcpx.JSONMarshaller{})

message2 := tcpx.NewURLPatternMessage("/login/",map[string]interface{}{"username":"tommy"})
buf2, _ := messag1.Pack(tcpx.JSONMarshaller{})

// conn.Write(buf1)
// conn.Write(buf2)
```