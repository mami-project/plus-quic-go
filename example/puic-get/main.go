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
var verbose = flag.Bool("verbose", true, "Verbose?")
var ofile = flag.String("ofile", "", "File to write statistics to")

func log(msg string, args ...interface{}) {
	if *verbose {
		fmt.Printf(msg, args...)
	}
}

func output(msg string, dataFile *os.File) {
	log(msg)
	if dataFile != nil {
		_, err := dataFile.WriteString(msg)

		if err != nil {
			panic(err)
		}
	}
}

func main() {
	flag.Parse()

	if *fetchURL == "" {
		panic("URL `%s' is not valid. Please specify an URL using -url.")
	}

	now := time.Now()

	var dst io.Writer
	var dataFile *os.File

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

	if *ofile != "" {
		f, err := os.OpenFile(*ofile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

		if err != nil {
			panic(err)
		}

		defer f.Close()

		dataFile = f
	}

	log("Fetching `%s' ...\n", *fetchURL)

	n, speed, elapsed, status, err := fetch(*fetchURL, dst)

	if err != nil {
		resultStr := fmt.Sprintf("%v;%d;%d;%f;%f;%d;%q\n", false, now.Unix(), n, speed, elapsed, status, err.Error())
		output(resultStr, dataFile)
		panic(err)
	}

	resultStr := fmt.Sprintf("%v;%d;%d;%f;%f;%d;%q\n", true, now.Unix(), n, speed, elapsed, status, "")
	output(resultStr, dataFile)
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
