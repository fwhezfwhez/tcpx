all-language-clients provide realizations of building tcpx's expected binary stream.

## All clients pack interface.
#### golang
golang pack has been realized, it can be referred.

https://github.com/fwhezfwhez/tcpx/blob/master/packx.go
```go
type Packx interface{
    Pack(messageID int32, src interface{}, header map[string]interface{}) ([]byte,error)
    UnPack([]byte, dest interface{}) (Message,error)
}
// example
func main() {
    type User struct{
        Username string `json:"username"`
    }
    // request payload
    var payload = User{"tcpx"}
    // packx
    packx := tcpx.NewPackx(tcpx.JsonMarshaller)
    // pack
    buf,_ :=packx.Pack(1, payload)
    // response payload
    var payload2 User
    packx.UnPack(buf, &payload2)
    // print {Username: "tcpx"}
    fmt.Println(payload2)
}
```

#### python
https://github.com/fwhezfwhez/tcpx/blob/master/all-language-clients/python/protocol.py
```python
    # payload
    payload = 'hello'

    # message
    message = TCPXMessage()
    message.id = 5
    message.header = {
        'header': '/tcpx/client1'
    }
    message.body = payload

    # tcpx instance
    tcpx_protocol = TCPXProtocol('json')

    # tcpx pack
    packed_data = tcpx_protocol.pack(message)
    
    # tcpx unpack
    message2 = TCPXMessage()
    payload2 = ''
    message2 = tcpx_protocol.unpack(packed_data, payload2)
    
    # print
    print(message2)
    print(payload2)
```

Validating server are provided too.

## Validating http program

```url
POST http://localhost:7000/tcpx/clients/stream/
application/json
```

payload:
```
<xml>
  <username>tcpx</username>
</xml>
```

**After Pack:**:  "AAAANAAAAAEAAAAEAAAAJG51bGw8eG1sPjx1c2VybmFtZT50Y3B4PC91c2VybmFtZT48L3htbD4="

request
```json
{
     "stream": "AAAANAAAAAEAAAAEAAAAJG51bGw8eG1sPjx1c2VybmFtZT50Y3B4PC91c2VybmFtZT48L3htbD4=",
     "marshal_name": "xml"
}
response
```json
{
    "message":"success",
    "ms":{"message_id":1,"header":null,"body":{"XMLName":{"Space":"","Local":"xml"},"Username":"tcpx"}},
    "result":"ok"
}

```

## 1. Run validating program
`cd all-language-clients`

`go run main.go`

## 2. Run clients
Before run clients, validating program should run first.
#### 2.1 go
`cd all-language-clients/go`

`go run main.go`

