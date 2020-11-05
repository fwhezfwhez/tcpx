package tcpx

import (
	"fmt"
	"testing"
)

func TestAnchor(t *testing.T) {
	mup := NewUrlPatternAnchor("/tcp-api/user-info/list-users/", 1)

	fmt.Println(Debug(mup))

	mi := NewMessageIDAnchor(10, 2)
	fmt.Println(Debug(mi))

}
