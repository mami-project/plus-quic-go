# wwwsrv

A simple H2QUIC server. It provides the methods:

* <prefix>256 : Return 256 random bytes
* <prefix>4KiB : Return 4KiB random bytes
* <prefix>1MiB : Return 1MiB random bytes
* <prefix>16MiB : Return 16MiB random bytes

Usage:

```
./wwwsrv -laddr=localhost:6121 -prefix=/data/
```
