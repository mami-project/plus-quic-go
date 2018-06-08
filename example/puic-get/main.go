package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
)

var fetchURL = flag.String("url", "", "URL to fetch.")
var path = flag.String("path", "", "Path to save file to.")

func log(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func main() {
	flag.Parse()

	if *fetchURL == "" {
		panic("URL `%s' is not valid. Please specify an URL using -url.")
	}

	var dst io.Writer

	if *path == "" {
		dst = &bytes.Buffer{}
	} else {
		fi, err := os.Create(*path)
		if err != nil {
			panic(err)
		}
		defer fi.Close()
		dst = fi
	}

	log("Fetching `%s' ...\n", *fetchURL)

	n, speed, elapsed, status, err := fetch(*fetchURL, dst)

	log("Bytes: %d, Speed: %f, Elapsed: %f, Status: %d, Err: %q\n", n, speed, elapsed, status, err)
}

func fetch(url string, dst io.Writer) (int64, float64, float64, int, error) {
	log("Creating h2client...\n")

	hclient := &http.Client{
		Transport: &h2quic.QuicRoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	start := time.Now()

	rsp, err := hclient.Get(url)

	if err != nil {
		return 0, 0.0, 0.0, -1, err
	}

	log("Status code: %d\n", rsp.StatusCode)
	log("Content length: %d\n", rsp.ContentLength)

	n, err := io.Copy(dst, rsp.Body)

	end := time.Now()
	elapsed := end.Sub(start).Seconds()

	speed := float64(n) / elapsed
	speed = speed / (1024.0 * 1024.0)

	log("Speed: %f MiB/s, elapsed: %f seconds\n", speed, elapsed)

	rsp.Body.Close()

	return n, speed, elapsed, rsp.StatusCode, err
}
