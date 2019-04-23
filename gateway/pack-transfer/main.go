// pack-transfer is used as a gateway to transfer a marshaled stream with specific messageID and optional header.
// 1. build and run the transfer server
//     If you'v configured go environment, then:
//         windows:
//             `go build main.go -o pack-transfer.exe`
//             cmd - `pack-transfer`
//         mac/linux:
//             `go build main.go -o pack-transfer`
//             command terminal - `./pack-transfer`
//
// If no go environment, then:
// windows: open cmd, cd dir-path,`pack-transfer-win64.exe`
// linux: `./pack-transfer-linux-64`
// mac: `./pack-transfer-mac-64`
//
// 2. send transfer request
//	 url: POST http://localhost:7000/gateway/pack/transfer/
//	 content-type: application/json
//	 body:
//		 {
//		    "marshal_name":<marshal_name>,
//          "stream": <marshaled_stream>,
//          "message_id": <message_id>
//          "header": <header>
//		 }
//   | arg_name | value range| type | necessary|
//   <marshal_name> | json,xml,protobuf,toml,yaml | string | yes|
//   <stream> | []byte... | []byte | yes|
//   <message_id> | 1,2,3,4…… | int32 | yes|
//   <header>   | {key:value, key2:value2}| map[string]interface{} | no|
package main

import (
	"flag"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"net/http"
	"time"
)

var port string

func init() {
	flag.StringVar(&port, "port", ":7000", "port, default :7000")
	flag.Parse()
}
func main() {
	r := gin.Default()
	r.POST("/gateway/pack/transfer/", func(c *gin.Context) {
		type Param struct {
			MarshalName string                 `json:"marshal_name" binding:"required"`
			Stream      []byte                 `json:"stream" binding:"required"`
			MessageID   int32                  `json:"message_id" binding:"required"`
			Header      map[string]interface{} `json:"header"`
		}
		var param Param
		if e := c.Bind(&param); e != nil {
			c.JSON(400, gin.H{"message": errorx.Wrap(e).Error()})
			return
		}
		marshaller, e := tcpx.GetMarshallerByMarshalName(param.MarshalName)
		if e != nil {
			c.JSON(400, gin.H{"message": errorx.Wrap(e).Error()})
			return
		}
		var packx = tcpx.NewPackx(marshaller)
		buf, e := packx.PackWithBody(param.MessageID, param.Stream, param.Header)
		if e != nil {
			c.JSON(400, gin.H{"message": errorx.Wrap(e).Error()})
			return
		}
		// valid whether correct
		var m interface{}
		messageInfo, e := packx.Unpack(buf, &m)
		if e != nil {
			c.JSON(500, gin.H{"message": errorx.Wrap(e).Error()})
			return
		}
		c.JSON(200, gin.H{"message": "success", "stream": buf, "message_info": messageInfo})
	})

	s := &http.Server{
		Addr:           port,
		Handler:        cors.AllowAll().Handler(r),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 21,
	}
	s.ListenAndServe()
}
