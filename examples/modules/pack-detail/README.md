## pack-detail
Here we learn how to encode/decode to a expected stream for tcpx.

#### introduction
Tcpx provides its official pack style --- messageID system. **This pack system supports all serializers user want to use.**.
Official tcpx lib provides `json`, `protobuf`, `toml`, `yaml`, `xml`. But you can design your own if necessary.


#### core
tcpx requires stream formated as:

```text
[4]byte -- length             fixed_size,binary big endian encode
[4]byte -- messageID          fixed_size,binary big endian encode
[4]byte -- headerLength       fixed_size,binary big endian encode
[4]byte -- bodyLength         fixed_size,binary big endian encode
[]byte -- header              marshal by json
[]byte -- body                marshal by marshaller
```
In any case, message stream should contains 6 parts.They are `length`, `messageID`, `headerLength`, `bodyLength`(It's ok if headerLength and bodyLength are 0. At this time, header and body are not consisted).

#### why design like this?
These Six parts are designed for solving problems below:

- **pack stuck**

**This is why design first part `length`.**

TCP stream are easily stuck to different message blocks.If one side sends two message blocks in a very short time, like [1 2 3 4]、[5,7,9].Another side will receive [1 2 3 4 5 7 9].You can't use it in this raw way.

To solve this, you can design message stream to a fixed length.This is a big waste because you must predict the bigest length of message.

Best solution is to add fixed byte to record a message's length.Then we profile the former raw stream into `[4 1 2 3 4]`, `[3 5 7 9]`.We can read the first byte and know how long a message is. However one byte is probably not big enough to record long message stream.Then extend it into 4 byte.

- **distinguish message type**

**This is why design second part `messageID`.**

When server side is receiving messages.They flow from a same place `net.Conn`.In http area, we know message is already clearly distinguished by url.But TCP/UDP/KCP doesn't have url.We must know which handle to handle which message.

To solve this, some people design `handler-name`.This might imitate url.This is using string to tell message type apart.It'ok.

Tcpx use int32 type `messageID` to tell.It's apparently smaller than string.Each messageID can point to a request or a response.

By int32 messageID, user knows which struct to receive in-come request.

Tcpx also use messageID to support middleware, which makes tcp used easy like http.

- **attach**

**This is why design third,fourth part `header`.**

In some cases, we want every request message with an atach.It might refer to a validating token, or a client property,or else.

We know HTTP has its request header too.HTTP saves `content-type`, `origin`...

If server requires client has a validating token, it's ugly to set it in payload.As below:

```json
{
    "payload": "",
	"token": "d8qj329siq912k33k5l2"
}
```

Payload should be clean:

```json
{ "payload":""}
```

So, tcpx designs a header part `headerLength`, `header`

- **payload**

**This is fifth,sixth part.**

payload is what a client really want to send.It can be well marshaled by json,protobuf,xml, etc.

Tcpx support all kinds of marshal way.


#### How to encode/decode?

**Encode**

Assume we want to send a message to server.It is `{"username":"tcpx"}`.And server require client attaches an auth.`{"auth":"abc"}`.Mark it messageID `1`.

First, we convert messageID into Second Part [0 0 0 1].This is using big-endian encoding.Every language has its lib to use big endian.In go:

```go
    var messageID = 1
    var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(message.MessageID))
	// messageIDBuf=[0 0 0 1]
```
**messageID**: `[0 0 0 1]`

Second, convert json header into bytes and count it length, convert it length to [4]byte via big endian:

**header**:`{"auth": "abc"}` -> `[123 34 97 117 116 104 34 58 34 97 98 99 34 125]`

length is 14 -> `[0 0 0 14]`

**headerLength**: `[0 0 0 14]`

```go
	var header []byte
	var headerLength = make([]byte, 4)

	var head = map[string]interface{}{
		"auth": "abc",
	}
	header, _ = json.Marshal(head)
	binary.BigEndian.PutUint32(headerLength, uint32(len(header)))
	// header = [123 34 97 117 116 104 34 58 34 97 98 99 34 125]
	// headerLength = [0 0 0 14]
```

Third, alike header, use your specific marshal way(json for example) to generate body and bodyLength.
**body**: `{"username":"tcpx"}` -> `[123 34 117 115 101 114 110 97 109 101 34 58 34 116 99 112 120 34 125]`

**bodyLength**: `19` -> `[0 0 0 19]`

```go
	var body []byte
	var bodyLength = make([]byte, 4)

	var payload = map[string]interface{}{
		"username": "tcpx",
	}
	body, _ = json.Marshal(payload)
	binary.BigEndian.PutUint32(bodyLength, uint32(len(body)))
	// body = [123 34 117 115 101 114 110 97 109 101 34 58 34 116 99 112 120 34 125]
	// bodyLength = [0 0 0 19]
```

Finally we count all length.
**length** `messageIDFixLength(4) + headerFixLength(4) + bodyFixedLength(4) + headerLength(14) + bodyLength(19)=45` -> `[0 0 0 45]`

Bind them in order of length, messageID, headerLength, bodyLength, header, body:
`
[0 0 0 45 0 0 0 1 0 0 0 14 0 0 0 19 123 34 97 117 116 104 34 58 34 97 98 99 34 125 123 34 117 115 101 114 110 97 109 101 34 58 34 116 99 112 120 34 125]`

All Done.

**Decode**

Assume we receive buf:
`
[0 0 0 45 0 0 0 1 0 0 0 14 0 0 0 19 123 34 97 117 116 104 34 58 34 97 98 99 34 125 123 34 117 115 101 114 110 97 109 101 34 58 34 116 99 112 120 34 125]`

Decode from bigendian:

length: buf[0:4] -> 45
messageID: buf[4:8] -> 1,
headerLength: buf[8:12] -> 14,
bodyLength: buf[12:16],

header: buf[16:16+headerLength] --->json unmarshal: {auth: abc}
body: buf[16+headerLength:16+headerLength+bodyLength] --> json unmarshal: {useranme: tcpx}

golang for example:

```go
	var buf = []byte{
		0, 0, 0, 45, 0, 0, 0, 1, 0, 0, 0, 14, 0, 0, 0, 19, 123, 34, 97, 117, 116, 104, 34, 58, 34, 97, 98, 99, 34, 125, 123, 34, 117, 115, 101, 114, 110, 97,
		109, 101, 34, 58, 34, 116, 99, 112, 120, 34, 125,
	}
    length := binary.BigEndian.Uint32(buf[0:4])
	// length = 45
	messageID := binary.BigEndian.Uint32(buf[4:8])
	// messageID = 1
	headerLength := binary.BigEndian.Uint32(buf[8:12])
	// headerLength = 14
	bodyLength := binary.BigEndian.Uint32(buf[12:16])
	// bodyLength = 19
    var header map[string]interface{}
	var body map[string]interface{}
	json.Unmarshal(buf[16:16+headerLength], &header)
	json.Unmarshal(buf[16+headerLength:16+headerLength+bodyLength], &body)
	// map[auth:abc]
    // map[username:tcpx]
```

#### Golang tcpx provides official api to decode and encode.

**encode**:
`import "github.com/fwhezfwhez/tcpx"`
```go
    buf,_ :=tcpx.PackWithMarshaller(Message{
		MessageID: 1,
		Header: map[string]interface{}{
			"auth": "abc",
		},
		Body: map[string]interface{}{
			"username":"tcpx",
		},
	}, JsonMarshaller{})
	// buf = [0 0 0 45 0 0 0 1 0 0 0 14 0 0 0 19 123 34 97 117 116 104 34 58 34 97 98 99 34 125 123 34 117 115 101 114 110 97 109 101 34 58 34 116 99 112 120 34 125]
```

**decode**:
`import "github.com/fwhezfwhez/tcpx"`
```go
	messageID,_ := tcpx.MessageIDOf(buf)
    header,_ := tcpx.HeaderOf(buf)
	bodyByte,_:= tcpx.BodyBytesOf(buf)
```

If you're aware of body struct, you can use like:
```go
    var u struct{Username string `json:"username"`}
    tcpx.UnpackWithMarshaller(buf, &u, tcpx.JsonMarshaller{})
```

