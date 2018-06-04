# puic-get

A tool to download a file through PLUS-enabled H2QUIC. 

```
puic-get -url https://example.com/example.txt -path /downloads/example.txt
```

The path option may be omitted in which case the data is downloaded to 'memory' only. 
If program terminates normally (see exit code) the last line contains CSV data of the form

```
successful (true/false);timestamp;speed in MiB/s;Time elapsed in seconds;HTTP Status Code;error message (if any)
```
