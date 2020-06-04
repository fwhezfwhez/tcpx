package tcpx

import (
	"io"
	"io/ioutil"
)

type Request struct{
    Body io.ReadCloser
    URL string
    Header map[string]interface{}
}

func NewRequest(url string, reader io.Reader) *Request{
	return &Request{
		Body: ioutil.NopCloser(reader),
		URL: url,
		Header: nil,
	}
}

func (r *Request) Set(key string, value interface{}) {
	if r.Header ==nil {
		r.Header = make(map[string]interface{},0)
	}

	r.Header[key] = value
}

