## marshaller
tcpx provides json,toml,yaml,protobuf marshaller, but if you want to use others,design it like:
```go
type OtherMarshaller struct{}
func (om OtherMarshaller) Marshal(v interface{}) ([]byte, error) {
    return []byte(""), nil
}
func (om OtherMarshaller) Unmarshal(data []byte, dest interface{}) error {
    return nil
}
func (om OtherMarshaller) MarshalName() string{
    return "other_marshaller"
}
```

server side to use your marshaller:
```go
srv.NewTcpX(OtherMarshaller)
```

client side to use your marshaller:
```go
message1 := tcpx.NewMessage(12, map[string]interface{}{"username":"tommy"})
buf1, _ := messag1.Pack(OtherMarshaller{})
```