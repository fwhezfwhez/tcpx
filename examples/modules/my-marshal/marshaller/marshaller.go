package marshaller

import "reflect"

type ByteMarshaller struct{}

func (om ByteMarshaller) Marshal(v interface{}) ([]byte, error) {
	return reflect.ValueOf(v).Bytes(), nil
}
func (om ByteMarshaller) Unmarshal(data []byte, dest interface{}) error {
	rv := reflect.ValueOf(dest)
	rv.Elem().Set(reflect.ValueOf(data))
	return nil
}
func (om ByteMarshaller) MarshalName() string {
	return "ByteMarshaller"
}
