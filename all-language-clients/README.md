all-language-clients provide realizations of building tcpx's expected binary stream.

## All clients pack interface.
#### golang
golang pack has been realized, it can be referred.
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
    packx := tcpx.NewPackx(tcpx.JsonMarshaller)
    buf,_ :=packx.Pack(1, User{"tcpx"})
    var us User
    packx.UnPack(buf, &us)
    // {Username: "tcpx"}
    fmt.Println(us)
}
```
```java
public interface Packx{

}
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

