# puic-poll

puic-poll takes a `;` separated list of URLs and randomly polls one of them and then waits for a specified amount
of seconds before polling. It writes statistics to a file as JSON and can send the statistics to a remote
REST API. 

```
Usage of ./puic-poll:
  -collect int
    	How many statistics items to collect before sending them. (default 10)
  -logfile string
    	File to write debug information to. (default "puic-poll.log")
  -send-to string
    	Where to send statistics to.
  -statsfile string
    	File to write statistics to. (default "puic-poll.csv")
  -urls string
    	URLs to fetch.
  -wait int
    	Time to wait in seconds before making the next request. (default 1)
```

If `send-to` is specified puic-poll will collect `collect` stats items before sending them as a POST request as `application/json`
to the URL provided in `send-to`. 


