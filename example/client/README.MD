# Client

In `quic-go/example/client`

```
/usr/local/go/bin/go run main.go https://localhost:6121/
```

This will completely crawl the target website, downloading everything but not writing
anything to disk.

Optionally you may provide an `-o=n` switch where `n` is the number of concurrent requests the crawler shall make. 

# Server

In `quic-go/example`

```
/usr/local/go/bin/go run main.go -www ~/
```

Replace `~/` with your `HTTP_ROOT`. The h2quic server will create an index if no index page is available so you can just point it to a random folder and then download all the files in it. 
