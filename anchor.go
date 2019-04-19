package tcpx

type MessageIDAnchor struct {
	MessageID   int32
	AnchorIndex int
}

type MiddlewareAnchor struct {
	MiddlewareKey string
	Middleware    func(c *Context)
	AnchorIndex   int
	ExpireAnchorIndex int
}
