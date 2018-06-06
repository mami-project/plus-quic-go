package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"bytes"
	"strings"
	"math/rand"
	"encoding/json"

	"github.com/lucas-clemente/quic-go/h2quic"
)

func openAppend(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return nil, err
	}

	return f, nil
}

func writeOrDie(dst io.Writer, msg string, args... interface{}) {
	now := time.Now().Unix()
	_, err := io.WriteString(dst, fmt.Sprintf("%d\t", now) + fmt.Sprintf(msg, args...) + "\n")

	if err != nil {
		panic(err)
	}
}

type Stats struct {
	Success bool
	Size int64
	Speed float64
	Elapsed float64
	StatusCode int
	Message string
	Now int64
}

func main() {
	var urls = flag.String("urls", "", "URLs to fetch.")
	var logfilePath = flag.String("logfile", "puic-poll.log", "File to write debug information to.")
	var statsfilePath = flag.String("statsfile", "puic-poll.csv", "File to write statistics to.")
	var wait = flag.Int("wait", 1, "Time to wait in seconds before making the next request.")
	var collect = flag.Int("collect", 10, "How many statistics items to collect before sending them.")
	var sendto = flag.String("send-to", "", "Where to send statistics to.")

	flag.Parse()

	var logfile *os.File
	var statsfile *os.File
	var err error
	
	collectedStats := make([]Stats, *collect)
	j := 0

	if *logfilePath != "" {
		logfile, err = openAppend(*logfilePath)

		if err != nil {
			panic(err)
		}

		
		defer logfile.Close()
	} else {
		logfile = os.Stdout
	}

	if *statsfilePath != "" {
		statsfile, err = openAppend(*statsfilePath)

		if err != nil {
			writeOrDie(logfile, fmt.Sprintf("ERR: Could not open statsfile $q: %q", statsfilePath, err.Error()))
			panic(err)
		}

		defer statsfile.Close()
	} else {
		statsfile = os.Stdout
	}

	urlsToFetch := strings.Split(*urls, ";")

	var netClient = &http.Client{
	  Timeout: time.Second * 5,
	}

	for {

		urlToFetch := urlsToFetch[rand.Int() % len(urlsToFetch)]

		size, speed, elapsed, statusCode, err := fetchOnce(urlToFetch, logfile)

		stats := Stats {
			Size : size,
			Speed : speed,
			Elapsed : elapsed,
			StatusCode : statusCode,
			Now : time.Now().Unix(),
		}

		if err != nil {
			stats.Success = false
			stats.Message = err.Error()
		} else {
			stats.Success = true
			stats.Message = ""
		}

		collectedStats[j] = stats

		statbytes, err := json.Marshal(stats)

		if err != nil {
			writeOrDie(logfile, fmt.Sprintf("ERR: Error: %q", err.Error()))
			panic(err)
		}

		statstr := string(statbytes)

		writeOrDie(statsfile, statstr)

		j++

		fmt.Println(j,*collect,*sendto)

		if j >= *collect {
			sendbytes, err := json.Marshal(collectedStats)

			if err != nil {
				writeOrDie(logfile, fmt.Sprintf("ERR: Error: %q", err.Error()))
				panic(err)
			}

			if *sendto != "" {
				writeOrDie(logfile, "Attempting to send...")
				sendTo(netClient, *sendto, sendbytes, logfile)
			}

			j = 0
		}

		time.Sleep(time.Duration(*wait) * time.Second)
	}
}

func sendTo(client *http.Client, url string, data []byte, log io.Writer) {
	_, err := client.Post(url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		writeOrDie(log, fmt.Sprintf("ERR: Could not send data to %q: %q", url, err.Error()))
	} else {
		writeOrDie(log, "Sent to %q", url)
	}
}

func fetchOnce(url string, log io.Writer) (int64, float64, float64, int, error) {
	writeOrDie(log, "Start fetching %q", url)

	hclient := &http.Client{
		Transport: &h2quic.QuicRoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	start := time.Now()

	rsp, err := hclient.Get(url)

	if err != nil {
		return 0, 0.0, 0.0, -1, err
	}

	writeOrDie(log, "Status code: %d", rsp.StatusCode)
	writeOrDie(log, "Content length: %d", rsp.ContentLength)

	n, err := io.Copy(&bytes.Buffer{}, rsp.Body)

	end := time.Now()
	elapsed := end.Sub(start).Seconds()

	speed := float64(n) / elapsed
	speed = speed / (1024.0 * 1024.0)

	writeOrDie(log, "Speed: %f MiB/s, elapsed: %f seconds", speed, elapsed)

	rsp.Body.Close()

	writeOrDie(log, "Done fetching %q", url)

	return n, speed, elapsed, rsp.StatusCode, err
}
