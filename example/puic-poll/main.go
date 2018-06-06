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
	_, err := io.WriteString(dst, fmt.Sprintf("%d\t", now) + fmt.Sprintf(msg, args...))

	if err != nil {
		panic(err)
	}
}

func main() {
	var urls = flag.String("urls", "", "URLs to fetch.")
	var logfilePath = flag.String("logfile", "puic-poll.log", "File to write debug information to.")
	var statsfilePath = flag.String("statsfile", "puic-poll.csv", "File to write statistics to.")
	var wait = flag.Int("wait", 1, "Time to wait in seconds before making the next request.")

	flag.Parse()

	var logfile *os.File
	var statsfile *os.File
	var err error

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
			writeOrDie(logfile, fmt.Sprintf("ERR: Could not open statsfile `%s': %q\n", statsfilePath, err.Error()))
			panic(err)
		}

		defer statsfile.Close()
	} else {
		statsfile = os.Stdout
	}

	urlsToFetch := strings.Split(*urls, ";")

	for {

		urlToFetch := urlsToFetch[rand.Int() % len(urlsToFetch)]

		size, speed, elapsed, statusCode, err := fetchOnce(urlToFetch, logfile)

		var statsLine string

		if err != nil {
			writeOrDie(logfile, fmt.Sprintf("ERR: Could not fetch %q: %q\n", urlToFetch, err.Error()))
			statsLine = fmt.Sprintf("ERR;%d;%f;%f;%d;%q\n", size, speed, elapsed, statusCode, err.Error())
		} else {
			statsLine = fmt.Sprintf("OK;%d;%f;%f;%d;%q\n", size, speed, elapsed, statusCode, "")
		}

		writeOrDie(statsfile, statsLine)

		time.Sleep(time.Duration(*wait) * time.Second)
	}
}

func fetchOnce(url string, log io.Writer) (int64, float64, float64, int, error) {
	writeOrDie(log, "Start fetching %q\n", url)

	hclient := &http.Client{
		Transport: &h2quic.QuicRoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	start := time.Now()

	rsp, err := hclient.Get(url)

	if err != nil {
		return 0, 0.0, 0.0, -1, err
	}

	writeOrDie(log, "Status code: %d\n", rsp.StatusCode)
	writeOrDie(log, "Content length: %d\n", rsp.ContentLength)

	n, err := io.Copy(&bytes.Buffer{}, rsp.Body)

	end := time.Now()
	elapsed := end.Sub(start).Seconds()

	speed := float64(n) / elapsed
	speed = speed / (1024.0 * 1024.0)

	writeOrDie(log, "Speed: %f MiB/s, elapsed: %f seconds\n", speed, elapsed)

	rsp.Body.Close()

	writeOrDie(log, "Done fetching `%s'\n", url)

	return n, speed, elapsed, rsp.StatusCode, err
}
