# puic-poll

puic-poll takes a `;` separated list of URLs and randomly polls one of them and then waits for a specified amount
of seconds before polling. It writes statistics to a file. Each line in the statistics file has the format

```
unixtimestamp<TAB>(ERR | OK);size;speed;elapsed;status code;error message
```

Example:

```
$ cat puic-poll.csv 
1528273755	OK;256;0.017063;0.014308;200;""
1528273760	OK;256;0.017713;0.013783;200;""
1528273765	OK;256;0.012648;0.019302;200;""
```

```
Usage of ./puic-poll:
  -logfile string
    	File to write debug information to. (default "puic-poll.log")
  -statsfile string
    	File to write statistics to. (default "puic-poll.csv")
  -urls string
    	URLs to fetch.
  -wait int
    	Time to wait in seconds before making the next request. (default 1)
```


