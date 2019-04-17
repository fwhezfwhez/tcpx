package tcpx

import (
	"encoding/binary"
	"reflect"
)

// a message stream consists of 4 parts:
// len: length of the message's messageID,header,body after packet
type Message struct {
	MessageID int32
	Header    map[string]interface{}
	Body      interface{}
}

// Get and Set don't have lock to ensure concurrently safe, which means
// if you should never operate the header in multiple goroutines, it's better to design a context yourself per request
// rather than straightly use message.Header.
func (msg Message) Get(key string) interface{} {
	if msg.Header == nil {
		return nil
	}
	return msg.Header[key]
}

// Get and Set don't have lock to ensure concurrently safe, which means
// if you should never operate the header in multiple goroutines, it's better to design a context yourself per request
// rather than straightly use message.Header.
func (msg *Message) Set(k string, v interface{}) {
	if msg.Header == nil {
		msg.Header[k] = v
	}
}

// PackWithMarshaller will encode message into blocks of length,messageID,headerLength,header,bodyLength,body.
// Users don't need to know how pack serializes itself if users use UnpackPWithMarshaller.
//
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// []byte -- header              marshal by marshaller
// [4]byte -- bodyLength         fixed_size,binary big endian encode
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
	var headerBuf []byte
	var bodyLengthBuf = make([]byte, 4)
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
	content = append(content, headerBuf...)
	content = append(content, bodyLengthBuf...)
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
// []byte -- header              marshal by marshaller
// [4]byte -- bodyLength         fixed_size,binary big endian encode
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
	// header
	var header map[string]interface{}
	e := marshaller.Unmarshal(stream[12:(12 + headerLength)], &header)
	if e != nil {
		return Message{}, e
	}
	// body长度
	bodyLength := binary.BigEndian.Uint32(stream[(12 + headerLength):(12 + headerLength + 4)])
	// body
	e = marshaller.Unmarshal(stream[(12 + headerLength + 4):(12 + headerLength + 4 + bodyLength)], dest)
	if e != nil {
		return Message{}, e
	}
	return Message{
		MessageID: int32(messageID),
		Header:    header,
		Body:      reflect.Indirect(reflect.ValueOf(dest)).Interface(),
	}, nil
}
