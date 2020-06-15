package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

func main() {
	var messageID = 1
	var header = map[string]interface{}{
		"auth":"abc",
	}
	var payload = map[string]interface{}{
		"username":"tcpx",
	}

	var e error
	//总长度
	var lengthBuf = make([]byte, 4)
	var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(messageID))
	var headerLengthBuf = make([]byte, 4)
	var bodyLengthBuf = make([]byte, 4)
	var headerBuf []byte
	var bodyBuf []byte
	// header 序列
	headerBuf, e = json.Marshal(header)
	if e != nil {
		panic(e)
	}
	// header长度的序列
	binary.BigEndian.PutUint32(headerLengthBuf, uint32(len(headerBuf)))
	if payload!=nil{
		// body 序列
		bodyBuf, e = json.Marshal(payload)
		if e != nil {
			panic(e)
		}
	}

	// body长度
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
}
