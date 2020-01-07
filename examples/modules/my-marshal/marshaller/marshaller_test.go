package marshaller

import (
	"fmt"
	"testing"
)

func TestMarshaler(t *testing.T) {
	mm := ByteMarshaller{}
	b, e := mm.Marshal([]byte(`hello`))
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	var rs []byte
	e = mm.Unmarshal(b, &rs)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	fmt.Println("rs:",string(rs))
}
