## Graceful

https://github.com/fwhezfwhez/tcpx/tree/master/examples/modules/graceful

- Graceful exit

Catch signal and do jobs arranged

- Graceful stop

two strategies:

1. `closeAllConnection = false` :Stop listen on, but no effect to existed connection

2. `closeAllConnection = true` :Stop listen on, stops all connection including connected clients.

- Graceful restart:

Contains `graceful stop` and `graceful start`. Between them, you can add jobs you want.
