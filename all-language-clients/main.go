// export http api to validate stream from all language clients
package main

import (
	"encoding/xml"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"net/http"
	"tcpx"
	"time"
)

func main() {
	r := gin.Default()
	r.POST("/tcpx/clients/stream/", func(c *gin.Context) {
		type Param struct {
			Stream      []byte `json:"stream"`
			MarshalName string `json:"marshal_name"`
		}
		var param Param
		e := c.Bind(&param)
		if e != nil {
			c.JSON(400, gin.H{"message": e.Error()})
			return
		}
		var user interface{}
		type JSONUser struct {
			Username string `json:"username"`
		}
		type XMLUser struct {
			XMLName  xml.Name `xml:"xml"`
			Username string   `xml:"username"`
		}
		type TOMLUser struct {
			Username string `toml:"username"`
		}
		type YAMLUser struct {
			Username string `yaml:"username"`
		}
		switch param.MarshalName {
		case "json":
			user = &JSONUser{}
		case "xml":
			user = &XMLUser{}
		case "toml", "tml":
			user = TOMLUser{}
		case "yaml", "yml":
			user = &YAMLUser{}
		case "protobuf", "proto":
			user = &User{}
		default:
			c.JSON(400, gin.H{"message": "marshal_name only accept ['json', 'xml', 'toml','yaml','protobuf']"})
			return
		}
		_, e = tcpx.UnpackWithMarshallerName(param.Stream, user, param.MarshalName)
		if e != nil {
			c.JSON(400, gin.H{"message": e.Error(), "result": "not ok"})
			return
		}
		c.JSON(200, gin.H{"message": "success", "result": "ok"})
	})
	s := &http.Server{
		Addr:           ":7000",
		Handler:        cors.AllowAll().Handler(r),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 21,
	}
	s.ListenAndServe()
}
