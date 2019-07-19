## middleware
Here we learn how to operate 3 kinds of middleware in server.

#### global-middleware
All handler will pass through.

`srv.UseGlobal(middleware)`

#### anchor-middleware
handler added after it's used.

```go
srv.Use("middleware-key", middleware)
srv.AddHandler(messageID, handler)
```

#### router-middleware
only specific handler will be affected

```go
srv.AddHandler(messageID, middleware, handler)
```

#### Order control

- Message flow passes in an order of `global-middleware` -> `anchor-middleware` -> `router-middleware`.
- To stop message flow, using `c.Abort()`.
- To execute or pass to next, using `c.Next()`.
