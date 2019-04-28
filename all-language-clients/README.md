all-language-clients provide realizations of building tcpx's expected binary stream.

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
```
## 1. Run validating program
`cd all-language-clients`

`go run main.go`

## 2. Run clients
Before run clients, validating program should run first.
#### 2.1 go
`cd all-language-clients/go`

`go run main.go`
