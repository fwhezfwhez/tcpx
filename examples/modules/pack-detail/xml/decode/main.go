package main

import (
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
)

func main() {
	var buf = []byte{
		0, 0, 0, 64, 0, 0, 0, 1, 0, 0, 0, 14, 0, 0, 0, 38, 123, 34, 97, 117, 116, 104, 34, 58, 34, 97, 98, 99, 34, 125, 60, 120, 109, 108, 62, 60, 117,
		115, 101, 114, 95, 110, 97, 109, 101, 62, 116, 99, 112, 120, 60, 47, 117, 115, 101, 114, 95, 110, 97, 109, 101, 62, 60, 47, 120, 109, 108, 62,
	}
	length := binary.BigEndian.Uint32(buf[0:4])
	messageID := binary.BigEndian.Uint32(buf[4:8])
	headerLength := binary.BigEndian.Uint32(buf[8:12])
	bodyLength := binary.BigEndian.Uint32(buf[12:16])

	var header map[string]interface{}
	var body struct {
		XMLName  xml.Name `xml:"xml"`
		Username string   `xml:"user_name"`
	}

	json.Unmarshal(buf[16:16+headerLength], &header)
	xml.Unmarshal(buf[16+headerLength:16+headerLength+bodyLength], &body)

	fmt.Println(length)
	fmt.Println(messageID)
	fmt.Println(headerLength)
	fmt.Println(bodyLength)

	fmt.Println(header)
	fmt.Println(body.Username)
}
