package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/fwhezfwhez/tcpx/examples/modules/pack-detail/protobuf/pb"
	"github.com/golang/protobuf/proto"
)

func main() {
	var messageID = 1
	var header = map[string]interface{}{
		"auth": "abc",
	}
	var payload = &pb.User{Username: "tcpx"}

	var e error
	var lengthBuf = make([]byte, 4)
	var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(messageID))
	var headerLengthBuf = make([]byte, 4)
	var bodyLengthBuf = make([]byte, 4)
	var headerBuf []byte
	var bodyBuf []byte
	headerBuf, e = json.Marshal(header)
	if e != nil {
		panic(e)
	}
	binary.BigEndian.PutUint32(headerLengthBuf, uint32(len(headerBuf)))
	if payload != nil {
		bodyBuf, e = proto.Marshal(payload)
		if e != nil {
			panic(e)
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

	fmt.Println(packet)
	// [0 0 0 32 0 0 0 1 0 0 0 14 0 0 0 6 123 34 97 117 116 104 34 58 34 97 98 99 34 125 10 4 116 99 112 120]
}
