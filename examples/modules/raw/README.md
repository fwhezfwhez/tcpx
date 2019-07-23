## raw
Here we learn how to send raw message.

#### detail
Somehow you don't like messageID block, and want to manage stream your own.You don't want to send stream with `header`, `messageID` ...
For example, you only want to send `[0 1]`

Here is how to receive and send message without designed stream.

#### step
`cd server`

`go run server.go`

`cd client`

`go run client.go`

#### output

```
before raw message in
use middleware 1
receive: hello,I am client.
```

```
hello,I am server.
```

#### notice

ListenRaw can't share the same port with others
