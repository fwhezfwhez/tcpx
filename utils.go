package tcpx

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"

	"github.com/fwhezfwhez/errorx"
)

type H map[string]interface{}

func Debug(src interface{}) string {
	buf, e := json.MarshalIndent(src, "  ", "  ")
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
	}
	return string(buf)
}

// Whether s in arr
// Support %%
func In(s string, arr []string) bool {
	for _, v := range arr {
		if strings.Contains(v, "%") {
			if strings.HasPrefix(v, "%") && strings.HasSuffix(v, "%") {
				if strings.Contains(s, string(v[1:len(v)-1])) {
					return true
				}
			} else if strings.HasPrefix(v, "%") {
				if strings.HasSuffix(s, string(v[1:])) {
					return true
				}
			} else if strings.HasSuffix(v, "%") {
				if strings.HasPrefix(s, string(v[:len(v)-1])) {
					return true
				}
			}
		} else {
			if v == s {
				return true
			}
		}
	}
	return false
}

// Defer eliminates all panic cases and handle panic reason by handlePanicError
func Defer(f func(), handlePanicError ...func(interface{})) {
	defer func() {
		if e := recover(); e != nil {
			if len(handlePanicError) == 0 {
				fmt.Printf("recover from %s\n", errorx.NewFromStringf("%v", e))
			}
			for _, handler := range handlePanicError {
				handler(e)
			}
		}
	}()
	f()
}

// CloseChanel(func(){close(chan)})
func CloseChanel(f func()) {
	defer func() {
		if e := recover(); e != nil {
			// when close(chan) panic from 'close of closed chan' do nothing
		}
	}()
	f()
}
func MD5(rawMsg string) string {
	data := []byte(rawMsg)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has)
	return strings.ToUpper(md5str1)
}

// Write full buf
// In case buf is too big and conn can't write once.
//
//
/*
   if len(buf)>65535 {
       connLock.Lock()
       WriteConn(buf, conn)
       connLock.Unlock()
    } else {
       conn.Write(buf)
   }
*/
//
func WriteConn(buf []byte, conn net.Conn) error {
	var sum = 0
	for {
		n, e := conn.Write(buf)

		if e != nil {
			if e == io.EOF {
				return io.EOF
			}
			return errorx.Wrap(e)
		}
		sum += n
		if sum >= len(buf) {
			break
		}
	}
	return nil
}

// TCPConnect will establish a tcp connection and return it
func TCPConnect(network string, url string) (net.Conn, error) {
	return net.Dial(network, url)
}

// WriteJSON will write conn a message wrapped by tcpx.JSONMarshaller
func WriteJSON(conn net.Conn, messageID int32, src interface{}) error {
	msg := Message{
		MessageID: messageID,
		Header:    nil,
		Body:      src,
	}

	buf, e := PackWithMarshaller(msg, JsonMarshaller{})
	if e != nil {
		return errorx.Wrap(e)
	}

	if _, e = conn.Write(buf); e != nil {
		return errorx.Wrap(e)
	}
	return nil
}

func BindJSON(bodyBuf []byte, dest interface{}) error {
	return json.Unmarshal(bodyBuf, dest)
}

func PipeJSON(conn net.Conn, args ...interface{}) error {

	if len(args) == 0 {
		return nil
	}

	if len(args)%2 != 0 {
		return errorx.NewFromString("iligal args, PipeJSON'args requires messageID int32, data interface{} in pair")
	}

	num := len(args) / 2

	var allBuf = make([]byte, 0, 100)

	for i := 0; i < len(args)-1; i += 2 {
		messageID, ok := args[i].(int)
		if !ok {
			return errorx.NewFromStringf("wrong type, args[%d] should be a int type messageID but got %s", i, reflect.TypeOf(args[i]).Name())
		}
		message := args[i+1]

		var msg Message
		if i == 0 {
			msg = Message{
				MessageID: int32(messageID),
				Header: map[string]interface{}{
					PIPED: fmt.Sprintf("enable;%d", num),
				},
				Body: message,
			}
		} else {
			msg = Message{
				MessageID: int32(messageID),
				Header:    nil,
				Body:      message,
			}
		}

		buf, e := PackWithMarshaller(msg, JsonMarshaller{})
		if e != nil {
			return errorx.Wrap(e)
		}

		allBuf = append(allBuf, buf...)
	}

	if len(allBuf) != 0 {
		if _, e := conn.Write(allBuf); e != nil {
			return errorx.Wrap(e)
		}
	}

	return nil
}

func TCPCallOnceJSON(network string, url string, messageID int, data interface{}) error {
	conn, e := TCPConnect(network, url)
	if e != nil {
		return errorx.Wrap(e)
	}
	defer conn.Close()

	msg := Message{
		MessageID: int32(messageID),
		Header:    nil,
		Body:      data,
	}
	buf, e := PackWithMarshaller(msg, JsonMarshaller{})
	if e != nil {
		return errorx.Wrap(e)
	}

	if _, e = conn.Write(buf); e != nil {
		return errorx.Wrap(e)
	}
	return nil
}

// get key-value from a header
func headerGetString(header map[string]interface{}, key string) (string, bool, error) {
	var exist bool
	var value string
	var valueI interface{}
	if len(header) == 0 {
		return "", false, nil
	}
	valueI, exist = header[key]

	if !exist {
		return "", exist, nil
	}

	var canConvert bool
	value, canConvert = valueI.(string)
	if !canConvert {
		return "", exist, errorx.NewFromStringf("key '%s'exist but is not a string type", key)
	}
	return value, exist, nil
}

// Recv a block of message from connection.
// To use this, it require sender sent message well packed by tcpx.Pack()
func Recv(conn net.Conn) (PackType, error) {
	return FirstBlockOf(conn)
}
