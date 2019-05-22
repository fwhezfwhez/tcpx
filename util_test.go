package tcpx

import (
	"fmt"
	"testing"
)

func TestIn(t *testing.T) {
	// true
	fmt.Println(In("ipskzk",[]string{"ip%"}))
	fmt.Println(In("ipskzk",[]string{"%ip%"}))
	fmt.Println(In("kjlk;lip",[]string{"%ip"}))
	fmt.Println(In("kjlk;lip",[]string{"%ip%"}))
}

func TestDebug(t *testing.T) {
	fmt.Println(Debug("hello"))
}

func TestDefer(t *testing.T) {
	f := func(){
		fmt.Println(1)
		panic(1)
	}
	Defer(f, func(v interface{}){
		fmt.Println(v)
	})
}
