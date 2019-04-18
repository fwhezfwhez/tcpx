package tcpx

import "encoding/json"

type Marshaller interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
	MarshalName() string
}
type JsonMarshaller struct{}

func (js JsonMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
func (js JsonMarshaller) Unmarshal(data []byte, dest interface{}) error {
	return json.Unmarshal(data, dest)
}

func (js JsonMarshaller) MarshalName() string{
	return "json"
}
