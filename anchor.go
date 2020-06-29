package tcpx

import (
	"fmt"
	"runtime"
	"strings"
)

type MessageIDAnchor struct {
	MessageID   int32
	AnchorIndex int
}

type AnchorMiddlewareInfo struct {
	WhereUse   []string
	WhereUnUse []string
}

type MiddlewareAnchor struct {
	MiddlewareKey string
	Middleware    func(c *Context)

	// anchorStartIndexRange.len should >= AnchorEndIndexRange.len, and should not bigger than 1.
	AnchorStartIndexRange []int
	AnchorEndIndexRange   []int

	info AnchorMiddlewareInfo
}

func (ma *MiddlewareAnchor) callUse(index int) {
	if len(ma.AnchorEndIndexRange) < len(ma.AnchorStartIndexRange) {
		panic(fmt.Errorf("can't Use(%s, handler) more than twice when the first anchor middleware has not called UnUse(%s) yet, they're at:\n%s",
			ma.MiddlewareKey, ma.MiddlewareKey, ma.FormatPath()))
	}

	// TODO, specific caller depth
	_, file, line, _ := runtime.Caller(2)

	ma.mergeUse(fmt.Sprintf("%s:%d", file, line))
	ma.AnchorStartIndexRange = append(ma.AnchorStartIndexRange, index)
}

func (ma *MiddlewareAnchor) callUnUse(index int) {
	if len(ma.AnchorEndIndexRange) >= len(ma.AnchorStartIndexRange) {
		panic(fmt.Errorf("can't UnUse(%s) when call UnUse times is bigger than call Use, they're at :\n%s",
			ma.MiddlewareKey, ma.FormatPath()))
	}

	// TODO, specific caller depth
	_, file, line, _ := runtime.Caller(2)
	ma.mergeUnUse(fmt.Sprintf("%s:%d", file, line))

	ma.AnchorEndIndexRange = append(ma.AnchorEndIndexRange, index)
}

func (ma *MiddlewareAnchor) mergeUse(path string) {
	if ma.info.WhereUse == nil {
		ma.info.WhereUse = make([]string, 0, 10)
	}

	ma.info.WhereUse = append(ma.info.WhereUse, path)
}

func (ma *MiddlewareAnchor) mergeUnUse(path string) {
	if ma.info.WhereUnUse == nil {
		ma.info.WhereUnUse = make([]string, 0, 10)
	}

	ma.info.WhereUnUse = append(ma.info.WhereUnUse, path)
}

func (ma *MiddlewareAnchor) FormatPath() string {
	var rs []string
	rs = append(rs, "------------------- Use -----------------------")

	rs = append(rs, ma.info.WhereUse...)
	rs = append(rs, "------------------- UnUse ---------------------")
	rs = append(rs, ma.info.WhereUnUse...)
	return strings.Join(rs, "\n")
}

func (ma *MiddlewareAnchor) Contains(handlerIndex int) bool {

	// When call same times Use and UnUse
	if len(ma.AnchorStartIndexRange) == len(ma.AnchorEndIndexRange) {
		for i := 0; i < len(ma.AnchorStartIndexRange); i++ {
			if ma.AnchorStartIndexRange[i] <= handlerIndex && ma.AnchorEndIndexRange[i] >= handlerIndex {
				return true
			}
		}
	} else {
		// When Use times is bigger by 1, first check last Use middleware
		if handlerIndex >= ma.AnchorStartIndexRange[len(ma.AnchorStartIndexRange)-1] {
			return true
		}

		for i := 0; i < len(ma.AnchorEndIndexRange); i ++ {
			if ma.AnchorStartIndexRange[i] <= handlerIndex && ma.AnchorEndIndexRange[i] >= handlerIndex {
				return true
			}
		}
	}

	return false
}

func (ma *MiddlewareAnchor) checkValidBeforeRun() {
	if len(ma.AnchorStartIndexRange) == len(ma.AnchorEndIndexRange) || len(ma.AnchorStartIndexRange) == len(ma.AnchorEndIndexRange)+1 {
		return
	}
	panic(fmt.Errorf("Use(%s, handler) times more than UnUse(%s) time by 2 times, it's fatal. You call them at:\n%s",
		ma.MiddlewareKey, ma.MiddlewareKey, ma.FormatPath()))
}
