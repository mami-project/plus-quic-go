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
	"path"
	"net"

	"github.com/lucas-clemente/quic-go/h2quic"
)

func openAppendOrDie(path string, log io.Writer) *os.File {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		writeOrDie(log, "ERR: Error %q", err.Error())
		panic(err)
	}

	return f
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
	Now int64
}

func main() {
	var urls = flag.String("urls", "", "URLs to fetch.")
	var logfilePath = flag.String("logfile", "puic-poll.log", "File to write debug information to.")
	var waitFrom = flag.Int("wait-from", 1000, "Minimum time to wait in milliseconds before making the next request.")
	var waitTo = flag.Int("wait-to", 5000, "Maximum time to wait in milliseconds before making the next request.")
	var collect = flag.Int("collect", 1024, "How many statistics items to collect in a single output file.")
	var odir = flag.String("odir", "./tmp/", "Output directory.")
	var ifaceName = flag.String("iface", "op0", "Interface to use.")

	flag.Parse()

	var logfile *os.File
	var ofile *os.File
	
	collectedStats := make([]Stats, *collect)
	fname := fmt.Sprintf("puic-poll-%d.json", time.Now().UnixNano())

	j := 0

	if *logfilePath != "" {
		logfile = openAppendOrDie(*logfilePath, nil)
		
		defer logfile.Close()
	} else {
		logfile = os.Stdout
	}

	iface, err := net.InterfaceByName(*ifaceName)

	if err != nil {
		writeOrDie(logfile, "ERR: Error using interface %q: %q", *ifaceName, err.Error())
		panic(err)
	}

	addrs, err := iface.Addrs()

	if err != nil {
		writeOrDie(logfile, "ERR: Error using interface %q: %q", *ifaceName, err.Error())
		panic(err)
	}

	if len(addrs) < 1 {
		writeOrDie(logfile, "ERR: Interface %q has no addresses?", *ifaceName)
		panic("Interface has no addresses")
	}

	ipAddr := addrs[0].(*net.IPNet).IP
	udpAddr := &net.UDPAddr { IP: ipAddr }

	writeOrDie(logfile, "Using %q", udpAddr.String())

	ofile = openAppendOrDie(path.Join(*odir, fname), logfile)

	urlsToFetch := strings.Split(*urls, ";")

	for {

		urlToFetch := urlsToFetch[rand.Int() % len(urlsToFetch)]

		size, speed, elapsed, statusCode, err := fetchOnce(urlToFetch, logfile)

		stats := Stats {
			Size : size,
			Speed : speed,
			Elapsed : elapsed,
			StatusCode : statusCode,
			Now : time.Now().UnixNano(),
		}

		if err != nil {
			stats.Success = false
			writeOrDie(logfile, fmt.Sprintf("ERR: Error: %q", err.Error()))
		} else {
			stats.Success = true
		}

		collectedStats[j] = stats

		statbytes, err := json.Marshal(stats)

		if err != nil {
			writeOrDie(logfile, fmt.Sprintf("ERR: Error: %q", err.Error()))
			panic(err)
		}

		statstr := string(statbytes)

		_, err = io.WriteString(ofile, statstr+"\n")

		if err != nil {
			writeOrDie(logfile, "ERR: Error writing to output file: %q", err.Error())
		}

		j++

		if j >= *collect {

			// Reset counter to zero, close the output file and open a new 
			// output file
			j = 0
			ofile.Close() 
			fname = fmt.Sprintf("puic-poll-%d.json", time.Now().UnixNano())
			ofile = openAppendOrDie(path.Join(*odir, fname), logfile)
		}

		wait := *waitFrom + (rand.Int() % (*waitTo - *waitFrom))

		time.Sleep(time.Duration(wait) * time.Millisecond)
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
