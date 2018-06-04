# puic-get

A tool to download a file through PLUS-enabled H2QUIC. 

```
./puic-get -url https://localhost:6121/mac.py -verbose=false -ofile=log.txt
```

The path option may be omitted in which case the data is downloaded to 'memory' only. 
If program terminates normally (see exit code) the last line contains CSV data of the form

```
successful (true/false);timestamp;speed in MiB/s;Time elapsed in seconds;HTTP Status Code;error message (if any)
```

or it can write it directly to a file (appending) (see `-ofile` option). 

Other options:

```
-ofile File to output results to
-verbose Verbose?
```
