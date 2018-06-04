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

or it can write it directly to a file (appending) (see `-ofile` option). `Successful` indicates whether there was some sort
of network error, `timestamp` indicates when the fetch was attempted as a unix timestamp, `elapsed` indicates how long it
took to download the file, `HTTP Status Code` simply contains the received response status code and `error message` contains
an error message should an error have occured while downloading the file. 

Other options:

```
-ofile File to output results to
-verbose Verbose?
```
