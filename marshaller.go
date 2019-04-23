package tcpx

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
)

type Marshaller interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
	MarshalName() string
}

func GetMarshallerByMarshalName(marshalName string) (Marshaller, error) {
	switch marshalName {
	case "json":
		return JsonMarshaller{}, nil
	case "xml":
		return XmlMarshaller{}, nil
	case "toml", "tml":
		return TomlMarshaller{}, nil
	case "yaml", "yml":
		return YamlMarshaller{}, nil
	case "protobuf", "proto":
		return ProtobufMarshaller{}, nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown marshalName %s,requires in [json,xml,toml,yaml,protobuf]", marshalName))
	}
}

type JsonMarshaller struct{}

func (js JsonMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
func (js JsonMarshaller) Unmarshal(data []byte, dest interface{}) error {
	return json.Unmarshal(data, dest)
}

func (js JsonMarshaller) MarshalName() string {
	return "json"
}

type XmlMarshaller struct{}

func (xm XmlMarshaller) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}
func (xm XmlMarshaller) Unmarshal(data []byte, dest interface{}) error {
	return xml.Unmarshal(data, dest)
}

func (xm XmlMarshaller) MarshalName() string {
	return "xml"
}

type YamlMarshaller struct{}

func (ym YamlMarshaller) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}
func (ym YamlMarshaller) Unmarshal(data []byte, dest interface{}) error {
	return yaml.Unmarshal(data, dest)
}

func (ym YamlMarshaller) MarshalName() string {
	return "yaml"
}

type TomlMarshaller struct{}

func (tm TomlMarshaller) Marshal(v interface{}) ([]byte, error) {
	return toml.Marshal(v)
}
func (tm TomlMarshaller) Unmarshal(data []byte, dest interface{}) error {
	return toml.Unmarshal(data, dest)
}

func (tm TomlMarshaller) MarshalName() string {
	return "toml"
}

type ProtobufMarshaller struct{}

func (pm ProtobufMarshaller) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

// dest should realize proto.Message
func (pm ProtobufMarshaller) Unmarshal(data []byte, dest interface{}) error {
	return json.Unmarshal(data, dest)
}

func (pm ProtobufMarshaller) MarshalName() string {
	return "protobuf"
}
