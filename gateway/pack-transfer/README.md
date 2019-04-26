gateway/pack-transfer provides http server  to build expected stream,supporting json,xml,toml,yaml,protobuf marshal types.

## start pack-transfer server
`cd ${GOPATH}/srv/tcpx`

`cd gateway/pack-transfer`

`go run main.go`

## detail
`go run main.go -port 7000`  run the gateway locally in port 7000 or else.

#### Gateway pack detail
**note: Each message should call once**
```url
POST http://localhost:7000/gateway/pack/transfer/
application/json
```
body:
```json
{
	"marshal_name":<marshal_name>,
	"stream": <stream>,
    "message_id": <message_id>,
    "header": <header>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| marshal_name | string |ranges in `"json","xml", "toml", "yaml", "protobuf"`| "json"|yes|
| stream | []byte | stream should be well marshalled by one of marshal_name | | yes|
|message_id | int32 | int32 type messageID| 1 | yes|
| header | map/object | key-value pairs | {"k1":"v1"}| no|
returns:
```json
{
    "message":<message>,
	"stream":<stream>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| message | string |"success" when status 200, "success", "error message" when 400/500 | "success"|yes|
| stream | []byte | packed stream,when error or status not 200, no stream field | | no|

example request:
payload:
```go
{"username": "hello, tcpx"}   ---json-->  "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0="
```
```json
{
    "marshal_name": "json",
    "stream": "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0=",
    "message_id": 1,
    "header": {
      "api": "/pack/"
    }
}
```
example response:
```json
{
    "stream": "AAAANgAAAAEAAAAQAAAAGnsiYXBpIjoiL3BhY2svIn17InVzZXJuYW1lIjoiaGVsbG8sIHRjcHgifQ=="
}
```
#### Gateway unpack detail
**note: able to unpack many messages once.**
```url
POST http://localhost:7000/gateway/unpack/transfer/
application/json
```
body:
```json
{
    "marshal_name": <marshal_name>,
	"stream": <stream>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| marshal_name | string |ranges in `"json","xml", "toml", "yaml", "protobuf"`| "json"|yes|
| stream | []byte | packed stream| | no|

returns:
```json
{
    "message": <message>,
	"blocks" <blocks>
}
```
| field | type | desc | example | nessessary
|---|---|--|--|--|
| message | string |"success" when status 200, "success", "error message" when 400/500 | "success"|yes|
| blocks | []block | unpacked blocks, when status not 200, no this field| | no|
|block| obj | each message block information, when status not 200,no this field | ++ look below++ | no|
block example:
```json
{
    "message_id": 1,
	"header": {"k1":"v1"},
	"marshal_name": "json",
	"stream": "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0="
}
```
example request:
```json
{
    "marshal_name": "json",
    "stream": "AAAANgAAAAEAAAAQAAAAGnsiYXBpIjoiL3BhY2svIn17InVzZXJuYW1lIjoiaGVsbG8sIHRjcHgifQ=="
}
```
example response:
```json
{
    "message": "success",
    "blocks": [
      {
        "message_id": 1,
        "header": {
          "api": "/pack/"
        },
        "marshal_name": "json",
        "stream": "eyJ1c2VybmFtZSI6ImhlbGxvLCB0Y3B4In0="
      }
    ]
}
```


