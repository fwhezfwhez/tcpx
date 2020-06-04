// export http api to validate stream from all language clients
package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/fwhezfwhez/tcpx"
	"github.com/fwhezfwhez/tcpx/all-language-clients/model"
	"github.com/rs/cors"
	"io/ioutil"
	"net/http"
	"time"
)
type H map[string]interface{}
type C struct{
	w http.ResponseWriter
	r *http.Request
}
func (c *C) Bind(dest interface{}) error {
	buf, e:= ioutil.ReadAll(c.r.Body)
	if e!=nil {
		return e
	}
	return json.Unmarshal(buf, dest)
}

func (c *C) JSON(statusCode int, data interface{}) {
	c.w.WriteHeader(statusCode)
	buf,_ := json.Marshal(data)
	c.w.Write(buf)
}
func main() {
	mux:= http.NewServeMux()
	mux.HandleFunc("/tcpx/clients/stream/", func(w http.ResponseWriter, r *http.Request) {
		var c = C{
			w:w,
			r:r,
		}

		type Param struct {
			Stream      []byte `json:"stream"`
			MarshalName string `json:"marshal_name"`
		}
		var param Param

		e := c.Bind(&param)
		if e != nil {
			c.JSON(400, H{"message": e.Error()})
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
			user = &TOMLUser{}
		case "yaml", "yml":
			user = &YAMLUser{}
		case "protobuf", "proto":
			user = &model.User{}
		default:
			c.JSON(400, H{"message": "marshal_name only accept ['json', 'xml', 'toml','yaml','protobuf']"})
			return
		}
		message, e := tcpx.UnpackWithMarshallerName(param.Stream, user, param.MarshalName)
		if e != nil {
			c.JSON(400, H{"message": e.Error(), "result": "not ok"})
			return
		}
		c.JSON(200, H{"message": "success", "result": "ok", "ms": message})
	})




	s := &http.Server{
		Addr:           ":7001",
		Handler:        cors.AllowAll().Handler(mux),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 21,
	}
	s.ListenAndServe()
}

func Debug(src interface{}) string {
	buf, _ := json.MarshalIndent(src, "  ", "  ")
	return string(buf)
}
