package tcpx

// Message contains the necessary parts of tcpx protocol
// MessagID is defining a message routing flag.
// Header is an attachment of a message.
// Body is the message itself, it should be raw message not serialized yet, like "hello", not []byte("hello")
type Message struct {
	MessageID int32                  `json:"message_id"`
	Header    map[string]interface{} `json:"header"`
	Body      interface{}            `json:"body"`
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
	msg.Header[k] = v
}