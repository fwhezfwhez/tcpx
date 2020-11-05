package tcpx

import (
	"encoding/json"
	"encoding/xml"
	"github.com/fwhezfwhez/errorx"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// PackType requires buffer message marshalled by tcpx.Pack
type PackType []byte

func (pt *PackType) BindJSON(dest interface{}) error {
	body, e := BodyBytesOf(*pt)
	if e != nil {
		return errorx.Wrap(e)
	}

	if e := json.Unmarshal(body, dest); e != nil {
		return errorx.Wrap(e)
	}
	return nil
}

func (pt *PackType) BindProtobuf(dest proto.Message) error {
	body, e := BodyBytesOf(*pt)
	if e != nil {
		return errorx.Wrap(e)
	}

	if e := proto.Unmarshal(body, dest); e != nil {
		return errorx.Wrap(e)
	}
	return nil
}
func (pt *PackType) BindTOML(dest interface{}) error {
	body, e := BodyBytesOf(*pt)
	if e != nil {
		return errorx.Wrap(e)
	}

	if e := UnmarshalTOML(body, dest); e != nil {
		return errorx.Wrap(e)
	}
	return nil
}
func (pt *PackType) BindYAML(dest interface{}) error {
	body, e := BodyBytesOf(*pt)
	if e != nil {
		return errorx.Wrap(e)
	}

	if e := yaml.Unmarshal(body, dest); e != nil {
		return errorx.Wrap(e)
	}
	return nil
}
func (pt *PackType) BindXML(dest interface{}) error {
	body, e := BodyBytesOf(*pt)
	if e != nil {
		return errorx.Wrap(e)
	}

	if e := xml.Unmarshal(body, dest); e != nil {
		return errorx.Wrap(e)
	}
	return nil
}

func (pt *PackType) URLPattern() (string, error) {
	urlPattern, e := URLPatternOf(*pt)

	if e != nil {
		return "", errorx.Wrap(e)
	}
	return urlPattern, nil
}

func (pt *PackType) MessageID() (int32, error) {
	msid, e := MessageIDOf(*pt)
	if e != nil {
		return msid, errorx.Wrap(e)
	}
	return msid, nil
}
