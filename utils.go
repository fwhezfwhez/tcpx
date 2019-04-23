package tcpx

import "encoding/json"

type H map[string]interface{}

func Debug(src interface{}) string {
	buf, _ := json.MarshalIndent(src, "  ", "  ")
	return string(buf)
}
