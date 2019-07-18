## graceful-exit
This example shows how to add jobs before app stop by outer signal accidentally.

#### step
- `go run main.go`
- ctrl + c

#### output
```go
receive signal: interrupt
prepare to stop server
clear online cache
job2 done
```
