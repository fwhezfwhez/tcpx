package tcpx

type TCPx struct {
	Marshaller Marshaller
}

// New a tcpx instance, specific a marshaller for communication.
// If marshaller is nil, official jsonMarshaller is put to used.
func NewTcpx(marshaller Marshaller) TCPx {
	if marshaller == nil {
		marshaller = JsonMarshaller{}
	}
	return TCPx{
		Marshaller: marshaller,
	}
}

// Pack src with specific messageID and optional headers
func (tcpx TCPx) Pack(messageID int32, src interface{}, headers ... map[string]interface{}) ([]byte, error) {
	if headers == nil || len(headers) == 0 {
		return PackWithMarshaller(Message{MessageID: messageID, Header: nil, Body: src}, tcpx.Marshaller)
	}
	var header = make(map[string]interface{}, 0)
	for _, v := range headers {
		for k1, v1 := range v {
			header [k1] = v1
		}
	}
	return PackWithMarshaller(Message{MessageID: messageID, Header: header, Body: src}, tcpx.Marshaller)
}
// Unpack
func (tcpx TCPx) Unpack(stream []byte, dest interface{}) (Message, error) {
	return UnpackWithMarshaller(stream, dest, tcpx.Marshaller)
}
