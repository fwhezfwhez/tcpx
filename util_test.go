package tcpx

import (
	"fmt"
	"testing"
)

func TestIn(t *testing.T) {
	// true
	fmt.Println(In("ipskzk", []string{"ip%"}))
	fmt.Println(In("ipskzk", []string{"%ip%"}))
	fmt.Println(In("kjlk;lip", []string{"%ip"}))
	fmt.Println(In("kjlk;lip", []string{"%ip%"}))
}

func TestDebug(t *testing.T) {
	fmt.Println(Debug("hello"))
}

func TestDefer(t *testing.T) {
	f := func() {
		fmt.Println(1)
		panic(1)
	}
	Defer(f, func(v interface{}) {
		fmt.Println(v)
	})
}

func TestMarshalToml(t *testing.T) {
	b, e := MarshalTOML(map[string]interface{}{
		"name": "ft",
	})
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	fmt.Println(string(b))

	var m map[string]interface{}

	if e := UnmarshalTOML(b, &m); e != nil {
		fmt.Println(e.Error())
		return
	}
	fmt.Println(m)
}
