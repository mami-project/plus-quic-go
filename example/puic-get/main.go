package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
)

var fetchURL = flag.String("url","","URL to fetch.")
var path = flag.String("path", "", "Path to save file to.")

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
		defer func() { fi.Close() }()
		dst = fi
	}

	fmt.Printf("Fetching `%s' ...\n", *fetchURL)
	err := fetch(*fetchURL, dst)

	if err != nil {
		panic(err)
	}
	fmt.Printf("Success!\n")
}

func fetch(url string, dst io.Writer) error {
	fmt.Printf("Creating h2client...\n")

	hclient := &http.Client{
		Transport: &h2quic.QuicRoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify:true}},
	}

	start := time.Now()

	rsp, err := hclient.Get(url)

	if err != nil {
		return err
	}



	fmt.Printf("Status code: %d\n", rsp.StatusCode)
	fmt.Printf("Content length: %d\n", rsp.ContentLength)
	
	n, err := io.Copy(dst, rsp.Body)


	end := time.Now()
	elapsed := end.Sub(start).Seconds()

	speed := float64(n)/elapsed

	fmt.Printf("Speed: %f MiB/s, elapsed: %f seconds\n", speed / (1024.0*1024.0), elapsed)

	rsp.Body.Close()

	return err
}

