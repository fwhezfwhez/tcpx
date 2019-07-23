package tcpx

import (
	"sync"
	"testing"
)

// Benchmark_PerRequest_OnMessage-4   	 2000000	       728 ns/op	    1336 B/op	       5 allocs/op
func Benchmark_PerRequest_OnMessage(b *testing.B) {
	srv := NewTcpX(nil)
	srv.OnMessage = func(c *Context) {
	}
	runTCPBench(b, srv)
}

// Benchmark_PerRequest_Mux-4   	 2000000	       832 ns/op	    1336 B/op	       5 allocs/op
func Benchmark_PerRequest_Mux(b *testing.B) {
	srv := NewTcpX(nil)
	srv.AddHandler(1, func(c *Context) {
	})
	runTCPBench(b, srv)
}

// Benchmark_PerRequest_Middleware-4   	 2000000	       870 ns/op	    1336 B/op	       5 allocs/op
func Benchmark_PerRequest_Middleware(b *testing.B) {
	srv := NewTcpX(nil)
	srv.UseGlobal(func(c *Context) {})
	srv.Use("mid-1", func(c *Context) {})
	srv.AddHandler(1, func(c *Context) {}, func(c *Context) {
	})
    runTCPBench(b, srv)
}

func runTCPBench(b *testing.B, srv *TcpX) {
	var l sync.Mutex
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			l.Lock()
			defer l.Unlock()
			ctx := &Context{Stream: PackStuff(1)}
			handleMiddleware(ctx, srv)
		}()
	}
}
