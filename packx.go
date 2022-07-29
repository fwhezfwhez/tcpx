package tcpx

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"io"
	"reflect"
)

// tcpx's tool to help build expected stream for communicating
type Packx struct {
	Marshaller Marshaller
}

// a package scoped packx instance
var packx = NewPackx(nil)
var PackJSON = NewPackx(JsonMarshaller{})
var PackTOML = NewPackx(TomlMarshaller{})
var PackXML = NewPackx(XmlMarshaller{})
var PackYAML = NewPackx(YamlMarshaller{})
var PackProtobuf = NewPackx(ProtobufMarshaller{})

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
// Src has not been marshaled yet.Whatever you put as src, it will be marshaled by packx.Marshaller.
func (packx Packx) Pack(messageID int32, src interface{}, headers ... map[string]interface{}) ([]byte, error) {
	if headers == nil || len(headers) == 0 {
		return PackWithMarshaller(Message{MessageID: messageID, Header: make(map[string]interface{}), Body: src}, packx.Marshaller)
	}
	var header = make(map[string]interface{}, 0)
	for _, v := range headers {
		for k1, v1 := range v {
			header [k1] = v1
		}
	}
	return PackWithMarshaller(Message{MessageID: messageID, Header: header, Body: src}, packx.Marshaller)
}

// PackWithBody is used for self design protocol
func (packx Packx) PackWithBody(messageID int32, body []byte, headers ...map[string]interface{}) ([]byte, error) {
	if headers == nil || len(headers) == 0 {
		return PackWithMarshallerAndBody(Message{MessageID: messageID, Header: make(map[string]interface{}), Body: nil}, body)
	}
	var header = make(map[string]interface{}, 0)
	for _, v := range headers {
		for k1, v1 := range v {
			header [k1] = v1
		}
	}
	return PackWithMarshallerAndBody(Message{MessageID: messageID, Header: header, Body: nil}, body)
}

// Unpack
// Stream is a block of length,messageID,headerLength,bodyLength,header,body.
// Dest refers to the body, it can be dynamic by messageID.
//
// Before use this, users should be aware of which struct used as `dest`.
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

// a stream from a reader can be apart by protocol.
// FirstBlockOf helps tear apart the first block []byte from reader
func (packx Packx) FirstBlockOf(r io.Reader) ([]byte, error) {
	return FirstBlockOf(r)
}
func (packx Packx) FirstBlockOfLimitMaxByte(r io.Reader, maxByte int32) ([]byte, error) {
	if maxByte <= 0 {
		return FirstBlockOf(r)
	}
	return FirstBlockOfLimitMaxByte(r, maxByte)
}

// returns the first block's messageID, header, body marshalled stream, error.
func UnPackFromReader(r io.Reader) (int32, map[string]interface{}, []byte, error) {
	buf, e := UnpackToBlockFromReader(r)
	if e != nil {
		return 0, nil, nil, e
	}

	messageID, e := MessageIDOf(buf)
	if e != nil {
		return 0, nil, nil, e
	}

	header, e := HeaderOf(buf)
	if e != nil {
		return 0, nil, nil, e
	}

	body, e := BodyBytesOf(buf)
	if e != nil {
		return 0, nil, nil, e
	}
	return messageID, header, body, nil
}

// Since FirstBlockOf has nothing to do with packx instance, so make it alone,
// for old usage remaining useful, old packx.FirstBlockOf is still useful
func FirstBlockOf(r io.Reader) ([]byte, error) {
	return UnpackToBlockFromReader(r)
}

func FirstBlockOfLimitMaxByte(r io.Reader, maxByte int32) ([]byte, error) {
	if maxByte <= 0 {
		return UnpackToBlockFromReader(r)
	}
	return UnpackToBlockFromReaderLimitMaxLengthOfByte(r, int(maxByte))
}

// a stream from a buffer which can be apart by protocol.
// FirstBlockOfBytes helps tear apart the first block []byte from a []byte buffer
func (packx Packx) FirstBlockOfBytes(buffer []byte) ([]byte, error) {
	return FirstBlockOfBytes(buffer)
}
func FirstBlockOfBytes(buffer []byte) ([]byte, error) {
	if len(buffer) < 16 {
		return nil, errors.New(fmt.Sprintf("require buffer length more than 16 but got %d", len(buffer)))
	}
	var length = binary.BigEndian.Uint32(buffer[0:4])
	if len(buffer) < 4+int(length) {
		return nil, errors.New(fmt.Sprintf("require buffer length more than %d but got %d", 4+int(length), len(buffer)))

	}
	return buffer[:4+int(length)], nil
}

// messageID of a stream.
// Use this to choose which struct for unpacking.
func (packx Packx) MessageIDOf(stream []byte) (int32, error) {
	return MessageIDOf(stream)
}

// messageID of a stream.
// Use this to choose which struct for unpacking.
func MessageIDOf(stream []byte) (int32, error) {
	if len(stream) < 8 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 8"))
	}
	messageID := binary.BigEndian.Uint32(stream[4:8])
	return int32(messageID), nil
}

// Length of the stream starting validly.
// Length doesn't include length flag itself, it refers to a valid message length after it.
func (packx Packx) LengthOf(stream []byte) (uint32, error) {
	return LengthOf(stream)
}

// Length of the stream starting validly.
// Length doesn't include length flag itself, it refers to a valid message length after it.
func LengthOf(stream []byte) (uint32, error) {
	if len(stream) < 4 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 4"))
	}
	length := binary.BigEndian.Uint32(stream[0:4])
	return length, nil
}

// Header length of a stream received
func (packx Packx) HeaderLengthOf(stream []byte) (int32, error) {
	return HeaderLengthOf(stream)
}

// Header length of a stream received
func HeaderLengthOf(stream []byte) (int32, error) {
	if len(stream) < 12 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 12"))
	}
	headerLength := binary.BigEndian.Uint32(stream[8:12])
	return int32(headerLength), nil
}

// Body length of a stream received
func (packx Packx) BodyLengthOf(stream []byte) (int32, error) {
	return BodyLengthOf(stream)
}

// Body length of a stream received
func BodyLengthOf(stream []byte) (int32, error) {
	if len(stream) < 16 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than %d", 16))
	}
	bodyLength := binary.BigEndian.Uint32(stream[12:16])
	return int32(bodyLength), nil
}

// Header bytes of a block
func (packx Packx) HeaderBytesOf(stream []byte) ([]byte, error) {
	return HeaderBytesOf(stream)
}

// Header bytes of a block
func HeaderBytesOf(stream []byte) ([]byte, error) {
	headerLen, e := HeaderLengthOf(stream)
	if e != nil {
		return nil, e
	}
	if len(stream) < 16+int(headerLen) {
		return nil, errors.New(fmt.Sprintf("stream lenth should be bigger than %d", 16+int(headerLen)))
	}
	header := stream[16 : 16+headerLen]
	return header, nil
}

// header of a block
func (packx Packx) HeaderOf(stream []byte) (map[string]interface{}, error) {
	return HeaderOf(stream)
}

// header of a block
func HeaderOf(stream []byte) (map[string]interface{}, error) {
	var header map[string]interface{}
	headerBytes, e := HeaderBytesOf(stream)
	if e != nil {
		return nil, errorx.Wrap(e)
	}
	e = json.Unmarshal(headerBytes, &header)
	if e != nil {
		return nil, errorx.Wrap(e)
	}
	return header, nil
}

// body bytes of a block
func (packx Packx) BodyBytesOf(stream []byte) ([]byte, error) {
	return BodyBytesOf(stream)
}

// body bytes of a block
func BodyBytesOf(stream []byte) ([]byte, error) {
	headerLen, e := HeaderLengthOf(stream)
	if e != nil {
		return nil, e
	}
	bodyLen, e := BodyLengthOf(stream)
	if e != nil {
		return nil, e
	}
	if len(stream) < 16+int(headerLen)+int(bodyLen) {
		return nil, errors.New(fmt.Sprintf("stream lenth should be bigger than %d", 16+int(headerLen)+int(bodyLen)))
	}
	body := stream[16+headerLen : 16+headerLen+bodyLen]
	return body, nil
}

// PackWithMarshaller will encode message into blocks of length,messageID,headerLength,header,bodyLength,body.
// Users don't need to know how pack serializes itself if users use UnpackPWithMarshaller.
//
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [0 0 0 24 0 0 0 1 0 0 0 6 0 0 0 6 2 1 19 18 13 11 11 3 1 23 12 132]
// header: [0 0 0 24]
// mesageID: [0 0 0 1]
// headerLength, bodyLength [0 0 0 6]
// header: [2 1 19 18 13 11]
// body: [11 3 1 23 12 132]
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// [4]byte -- bodyLength         fixed_size,binary big endian encode
// []byte -- header              marshal by json
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
	headerBuf, e = json.Marshal(message.Header)
	if e != nil {
		return nil, e
	}
	binary.BigEndian.PutUint32(headerLengthBuf, uint32(len(headerBuf)))
	if message.Body != nil {
		bodyBuf, e = marshaller.Marshal(message.Body)
		if e != nil {
			return nil, e
		}
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

// same as above
func PackWithMarshallerName(message Message, marshallerName string) ([]byte, error) {
	var marshaller Marshaller
	switch marshallerName {
	case "json":
		marshaller = JsonMarshaller{}
	case "xml":
		marshaller = XmlMarshaller{}
	case "toml", "tml":
		marshaller = TomlMarshaller{}
	case "yaml", "yml":
		marshaller = YamlMarshaller{}
	case "protobuf", "proto":
		marshaller = ProtobufMarshaller{}
	default:
		return nil, errors.New("only accept ['json', 'xml', 'toml','yaml','protobuf']")
	}
	return PackWithMarshaller(message, marshaller)
}

// unpack stream from PackWithMarshaller
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// [4]byte -- bodyLength         fixed_size,binary big endian encode
// []byte -- header              marshal by json
// []byte -- body                marshal by marshaller
func UnpackWithMarshaller(stream []byte, dest interface{}, marshaller Marshaller) (Message, error) {
	if marshaller == nil {
		marshaller = JsonMarshaller{}
	}
	var e error
	// 包长
	length := binary.BigEndian.Uint32(stream[0:4])
	stream = stream[0 : length+4]
	// messageID
	messageID := binary.BigEndian.Uint32(stream[4:8])
	// header长度
	headerLength := binary.BigEndian.Uint32(stream[8:12])
	// body长度
	bodyLength := binary.BigEndian.Uint32(stream[12:16])
	// header
	var header map[string]interface{}
	if headerLength != 0 {
		e = json.Unmarshal(stream[16:(16 + headerLength)], &header)
		if e != nil {
			return Message{}, e
		}
	}

	// body
	if bodyLength != 0 {
		e = marshaller.Unmarshal(stream[16+headerLength:(16 + headerLength + bodyLength)], dest)
		if e != nil {
			return Message{}, e
		}
	}

	return Message{
		MessageID: int32(messageID),
		Header:    header,
		Body:      reflect.Indirect(reflect.ValueOf(dest)).Interface(),
	}, nil
}

// same as above
func UnpackWithMarshallerName(stream []byte, dest interface{}, marshallerName string) (Message, error) {
	var marshaller Marshaller
	switch marshallerName {
	case "json":
		marshaller = JsonMarshaller{}
	case "xml":
		marshaller = XmlMarshaller{}
	case "toml", "tml":
		marshaller = TomlMarshaller{}
	case "yaml", "yml":
		marshaller = YamlMarshaller{}
	case "protobuf", "proto":
		marshaller = ProtobufMarshaller{}
	default:
		return Message{}, errors.New("only accept ['json', 'xml', 'toml','yaml','protobuf']")
	}
	return UnpackWithMarshaller(stream, dest, marshaller)
}

// unpack the first block from the reader.
// protocol is PackWithMarshaller().
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// [4]byte -- bodyLength         fixed_size,binary big endian encode
// []byte -- header              marshal by json
// []byte -- body                marshal by marshaller
// ussage:
// for {
//     blockBuf, e:= UnpackToBlockFromReader(reader)
// 	   go func(buf []byte){
//         // handle a message block apart
//     }(blockBuf)
//     continue
// }
func UnpackToBlockFromReader(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("reader is nil")
	}
	var info = make([]byte, 4, 4)
	if e := readUntil(reader, info); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	length, e := packx.LengthOf(info)
	if e != nil {
		return nil, e
	}
	var content = make([]byte, length, length)
	if e := readUntil(reader, content); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	return append(info, content ...), nil
}

func UnpackToBlockFromReaderLimitMaxLengthOfByte(reader io.Reader, maxByTe int) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("reader is nil")
	}
	var info = make([]byte, 4, 4)
	if e := readUntil(reader, info); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	length, e := packx.LengthOf(info)
	if e != nil {
		return nil, e
	}

	if length > uint32(maxByTe) {
		return nil, errorx.NewFromStringf("recv message beyond max byte length limit(%d), got (%d)", maxByTe, length)
	}

	var content = make([]byte, length, length)
	if e := readUntil(reader, content); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	return append(info, content ...), nil
}

func readUntil(reader io.Reader, buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	var offset int
	for {
		n, e := reader.Read(buf[offset:])
		if e != nil {
			if e == io.EOF {
				return e
			}
			return errorx.Wrap(e)
		}
		offset += n
		if offset >= len(buf) {
			break
		}
	}
	return nil
}

// This method is used to pack message whose body is well-marshaled.
func PackWithMarshallerAndBody(message Message, body []byte) ([]byte, error) {
	var e error
	var lengthBuf = make([]byte, 4)
	var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(message.MessageID))
	var headerLengthBuf = make([]byte, 4)
	var bodyLengthBuf = make([]byte, 4)
	var headerBuf []byte
	var bodyBuf []byte
	headerBuf, e = json.Marshal(message.Header)
	if e != nil {
		return nil, e
	}
	binary.BigEndian.PutUint32(headerLengthBuf, uint32(len(headerBuf)))
	bodyBuf = body
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

func PackHeartbeat() []byte {
	buf, e := PackWithMarshallerAndBody(Message{
		MessageID: DEFAULT_HEARTBEAT_MESSAGEID,
	}, nil)
	if e != nil {
		panic(e)
	}
	return buf
}

// pack short signal which only contains messageID
func PackStuff(messageID int32) []byte {
	buf, e := PackWithMarshallerAndBody(Message{
		MessageID: messageID,
	}, nil)
	if e != nil {
		panic(e)
	}
	return buf
}

func URLPatternOf(stream []byte) (string, error) {
	header, e := HeaderOf(stream)
	if e != nil {
		return "", errorx.Wrap(e)
	}
	str, _, e := headerGetString(header, HEADER_ROUTER_VALUE)
	if e != nil {
		return "", errorx.Wrap(e)
	}
	return str, nil
}

func RouteTypeOf(stream []byte) (string, error) {
	header, e := HeaderOf(stream)
	if e != nil {
		return "", errorx.Wrap(e)
	}
	str, _, e := headerGetString(header, HEADER_ROUTER_KEY)
	if e != nil {
		return "", errorx.Wrap(e)
	}

	return str, nil
}

// pack detail
func Pack(messageID int32, header map[string]interface{}, src interface{}, marshaller Marshaller) ([]byte, error) {
	return PackWithMarshaller(Message{
		MessageID: messageID,
		Header:    header,
		Body:      src,
	}, marshaller)
}
