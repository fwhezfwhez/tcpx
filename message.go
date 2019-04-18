package tcpx


type Message struct {
	MessageID int32
	Header    map[string]interface{}
	Body      interface{}
}

// Get value of message's header whose key is 'key'
// Get and Set don't have lock to ensure concurrently safe, which means
// if you should never operate the header in multiple goroutines, it's better to design a context yourself per request
// rather than straightly use message.Header.
func (msg Message) Get(key string) interface{} {
	if msg.Header == nil {
		return nil
	}
	return msg.Header[key]
}

// Get and Set don't have lock to ensure concurrently safe, which means
// if you should never operate the header in multiple goroutines, it's better to design a context yourself per request
// rather than straightly use message.Header.
func (msg *Message) Set(k string, v interface{}) {
	if msg.Header == nil {
		msg.Header[k] = v
	}
}

