package tcpx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

type Packx struct {
	Marshaller Marshaller
}



// New a packx instance, specific a marshaller for communication.
// If marshaller is nil, official jsonMarshaller is put to used.
func NewPackx(marshaller Marshaller) *Packx {
	if marshaller == nil {
		marshaller = JsonMarshaller{}
	}
	return &Packx{
		Marshaller: marshaller,
	}
}

// Pack src with specific messageID and optional headers
func (packx Packx) Pack(messageID int32, src interface{}, headers ... map[string]interface{}) ([]byte, error) {
	if headers == nil || len(headers) == 0 {
		return PackWithMarshaller(Message{MessageID: messageID, Header: nil, Body: src}, packx.Marshaller)
	}
	var header = make(map[string]interface{}, 0)
	for _, v := range headers {
		for k1, v1 := range v {
			header [k1] = v1
		}
	}
	return PackWithMarshaller(Message{MessageID: messageID, Header: header, Body: src}, packx.Marshaller)
}

// Unpack
// Before use this users should be aware of which struct should be used as `dest`.
// You can use stream's messageID for judgement like:
// messageID,_:= packx.MessageIDOf(stream)
// switch messageID {
//     case 1:
//       packx.Unpack(stream, &struct1)
//     case 2:
//       packx.Unpack(stream, &struct2)
//     ...
// }
func (packx Packx) Unpack(stream []byte, dest interface{}) (Message, error) {
	return UnpackWithMarshaller(stream, dest, packx.Marshaller)
}

// messageID of a stream.
// Use this to choose which struct for unpacking.
func (packx Packx) MessageIDOf(stream []byte) (int32, error) {
	if len(stream) < 8 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 8"))
	}
	messageID := binary.BigEndian.Uint32(stream[4:8])
	return int32(messageID), nil
}

// Length of the stream starting validly.
// Length doesn't include length flag itself, it refers to a valid message length after it.
func (packx Packx) LengthOf(stream []byte) (int32, error) {
	if len(stream) < 4 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 4"))
	}
	length := binary.BigEndian.Uint32(stream[0:4])
	return int32(length), nil
}

func (packx Packx) HeaderLengthOf(stream []byte) (int32, error) {
	if len(stream) < 12 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 12"))
	}
	headerLength := binary.BigEndian.Uint32(stream[8:12])
	return int32(headerLength), nil
}
func (packx Packx) BodyLengthOf(stream []byte) (int32, error) {
	if len(stream) < 16 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than %d", 16))
	}
	bodyLength := binary.BigEndian.Uint32(stream[12:16])
	return int32(bodyLength), nil
}


// PackWithMarshaller will encode message into blocks of length,messageID,headerLength,header,bodyLength,body.
// Users don't need to know how pack serializes itself if users use UnpackPWithMarshaller.
//
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// [4]byte -- bodyLength         fixed_size,binary big endian encode
// []byte -- header              marshal by marshaller
// []byte -- body                marshal by marshaller
func PackWithMarshaller(message Message, marshaller Marshaller) ([]byte, error) {
	if marshaller == nil {
		marshaller = JsonMarshaller{}
	}
	var e error
	var lengthBuf = make([]byte, 4)
	var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(message.MessageID))
	var headerLengthBuf = make([]byte, 4)
	var bodyLengthBuf = make([]byte, 4)
	var headerBuf []byte
	var bodyBuf []byte
	headerBuf, e = marshaller.Marshal(message.Header)
	if e != nil {
		return nil, e
	}
	binary.BigEndian.PutUint32(headerLengthBuf, uint32(len(headerBuf)))
	bodyBuf, e = marshaller.Marshal(message.Body)
	if e != nil {
		return nil, e
	}
	binary.BigEndian.PutUint32(bodyLengthBuf, uint32(len(bodyBuf)))
	var content = make([]byte, 0, 1024)

	content = append(content, messageIDBuf...)
	content = append(content, headerLengthBuf...)
	content = append(content, bodyLengthBuf...)
	content = append(content, headerBuf...)
	content = append(content, bodyBuf...)

	binary.BigEndian.PutUint32(lengthBuf, uint32(len(content)))

	var packet = make([]byte, 0, 1024)

	packet = append(packet, lengthBuf...)
	packet = append(packet, content...)
	return packet, nil
}

// unpack stream from PackWithMarshaller
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// [4]byte -- bodyLength         fixed_size,binary big endian encode
// []byte -- header              marshal by marshaller
// []byte -- body                marshal by marshaller
func UnpackWithMarshaller(stream []byte, dest interface{}, marshaller Marshaller) (Message, error) {
	if marshaller == nil {
		marshaller = JsonMarshaller{}
	}
	// 包长
	length := binary.BigEndian.Uint32(stream[0:4])
	stream = stream[0:length]
	// messageID
	messageID := binary.BigEndian.Uint32(stream[4:8])
	// header长度
	headerLength := binary.BigEndian.Uint32(stream[8:12])
	// body长度
	bodyLength := binary.BigEndian.Uint32(stream[12:16])
	// header
	var header map[string]interface{}
	e := marshaller.Unmarshal(stream[16:(16 + headerLength)], &header)
	if e != nil {
		return Message{}, e
	}
	// body
	e = marshaller.Unmarshal(stream[16+headerLength:(16 + headerLength + bodyLength)], dest)
	if e != nil {
		return Message{}, e
	}
	return Message{
		MessageID: int32(messageID),
		Header:    header,
		Body:      reflect.Indirect(reflect.ValueOf(dest)).Interface(),
	}, nil
}

