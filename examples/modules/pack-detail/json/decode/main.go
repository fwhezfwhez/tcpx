package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

func main() {
	var buf = []byte{
		0, 0, 0, 45, 0, 0, 0, 1, 0, 0, 0, 14, 0, 0, 0, 19, 123, 34, 97, 117, 116, 104, 34, 58, 34, 97, 98, 99, 34, 125, 123, 34, 117, 115, 101, 114, 110, 97,
		109, 101, 34, 58, 34, 116, 99, 112, 120, 34, 125,
	}
	length := binary.BigEndian.Uint32(buf[0:4])
	messageID := binary.BigEndian.Uint32(buf[4:8])
	headerLength := binary.BigEndian.Uint32(buf[8:12])
	bodyLength := binary.BigEndian.Uint32(buf[12:16])

	var header map[string]interface{}
	var body map[string]interface{}

	json.Unmarshal(buf[16:16+headerLength], &header)
	json.Unmarshal(buf[16+headerLength:16+headerLength+bodyLength], &body)

	fmt.Println(length)
	fmt.Println(messageID)
	fmt.Println(headerLength)
	fmt.Println(bodyLength)

	fmt.Println(header)
	fmt.Println(body)
}
