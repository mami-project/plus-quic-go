# puic-poll

puic-poll takes a `;` separated list of URLs and randomly polls one of them and then waits for a specified amount
of seconds before polling. It writes statistics to a file as JSON and can send the statistics to a remote
REST API. 

```
Usage of ./puic-poll:
  -certs string
    	Path to certificates to be trusted as Root CAs.
  -collect int
    	How many statistics items to collect in a single output file. (default 1024)
  -iface string
    	Interface to use. (default "op0")
  -logfile string
    	File to write debug information to. (default "puic-poll.log")
  -odir string
    	Output directory. (default "./tmp/")
  -urls string
    	URLs to fetch (; delimited).
  -wait-from int
    	Minimum time to wait in milliseconds before making the next request. (default 1000)
  -wait-to int
    	Maximum time to wait in milliseconds before making the next request. (default 5000)

```

Output files have the name `puic-poll-<UNIXNANO>.json`. If logfile is an empty string stdout is used instead.

Example Usage:

```
./puic-poll -certs=rootCACert.pem -collect=10 -iface=lo -logfile="" -odir="./tmp/" -urls=https://localhost:6121/data/256 -wait-from=500 -wait-to=1000
```
