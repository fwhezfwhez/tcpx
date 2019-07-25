package tcpx

import (
	"testing"
)

// Benchmark_PerRequest_OnMessage-4   	 2000000	       643 ns/op	    1368 B/op	       5 allocs/op
func Benchmark_PerRequest_OnMessage(b *testing.B) {
	srv := NewTcpX(nil)
	srv.OnMessage = func(c *Context) {
	}
	runTCPBench(b, srv)
}

// Benchmark_PerRequest_Mux-4   	 2000000	       761 ns/op	    1368 B/op	       5 allocs/op
func Benchmark_PerRequest_Mux(b *testing.B) {
	srv := NewTcpX(nil)
	srv.AddHandler(1, func(c *Context) {
	})
	runTCPBench(b, srv)
}

// Benchmark_PerRequest_Middleware-4   	 2000000	       768 ns/op	    1368 B/op	       5 allocs/op
func Benchmark_PerRequest_Middleware(b *testing.B) {
	srv := NewTcpX(nil)
	srv.UseGlobal(func(c *Context) {})
	srv.Use("mid-1", func(c *Context) {})
	srv.AddHandler(1, func(c *Context) {}, func(c *Context) {
	})
    runTCPBench(b, srv)
}

func runTCPBench(b *testing.B, srv *TcpX) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			ctx := &Context{Stream: PackStuff(1)}
			handleMiddleware(ctx, srv)
		}()
	}
}
