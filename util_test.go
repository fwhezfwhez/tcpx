package tcpx

import (
	"fmt"
	"testing"
)

func TestIn2(t *testing.T) {
	// true
	fmt.Println(In("ipskzk",[]string{"ip%"}))
	fmt.Println(In("ipskzk",[]string{"%ip%"}))
	fmt.Println(In("kjlk;lip",[]string{"%ip"}))
	fmt.Println(In("kjlk;lip",[]string{"%ip%"}))
}
