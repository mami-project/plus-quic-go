# puic-get

A tool to download a file through PLUS-enabled H2QUIC. 

Usage:

```
Usage of ./puic-get:
  -path string
    	Path to save file to.
  -url string
    	URL to fetch.
```

If no path is specified it'll download to stdout.


Example Usage:

```
./puic-get -url=https://localhost:6121/data/256
```


