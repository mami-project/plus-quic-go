package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"bytes"
	"strings"
	"math/rand"
	"encoding/json"
	"path"
	"net"
	"crypto/x509"

	"github.com/lucas-clemente/quic-go/h2quic"
)

// opens a file in append,wronly,create,0600 mode
// or panics if that's not possible.
func openAppendOrDie(path string, log io.Writer) *os.File {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		writeOrDie(log, "ERR: Error %q", err.Error())
		panic(err)
	}

	return f
}

// writes a msg (fmt) with args or panics if that's not
// possible.
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

// Create a HTTP Client using an H2Quic QuicRoundTripper that determines
// the LocalAddr to listen on based on the iface name provided. 
// (it'll pick the first address of that interface). If the interface name
// provided is an empty string it'll listen on the zero address. 
func createHttpClient(ifaceName string, log io.Writer) (*http.Client, error) {
	if ifaceName == "" {
		hclient := &http.Client{
			Transport: &h2quic.QuicRoundTripper{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		return hclient, nil
	}

	// Figure out the ip address of the specified interface
	// then pick the first one later on to use as LocalAddr
	iface, err := net.InterfaceByName(ifaceName)

	if err != nil {
		return nil, fmt.Errorf("ERR: Error using interface %q: %q", ifaceName, err.Error())
	}

	addrs, err := iface.Addrs()

	if err != nil {
		return nil, fmt.Errorf("ERR: Error using interface %q: %q", ifaceName, err.Error())
	}

	if len(addrs) < 1 {
		return nil, fmt.Errorf("ERR: Interface %q has no addresses?", ifaceName)
	}

	ipAddr := addrs[0].(*net.IPNet).IP
	udpAddr := &net.UDPAddr { IP: ipAddr }

	writeOrDie(log, "Using %q", udpAddr.String())

	hclient := &http.Client{
		Transport: &h2quic.QuicRoundTripper{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			LocalAddr : udpAddr,
		},
	}

	return hclient, nil
}

// Opens the logfile. If the path is an empty string
// it'll point to stdout. Panics if the logfile can not be opened.
func openLogfile(logfilePath string) *os.File {
	if logfilePath != "" {
		logfile := openAppendOrDie(logfilePath, nil)

		return logfile
	} else {
		return os.Stdout
	}
}

// Return the name of the next output file name
func getOFileName() string {
	return fmt.Sprintf("puic-poll-%d.json", time.Now().UnixNano())
}

// Open the next output file in append,wronly,create,0600 mode.
func openNextOutputFile(odir string) (*os.File, error) {
	fpath := path.Join(odir, getOFileName())

	f, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return nil, err
	} else {
		return f, nil
	}
}

// Sleep for at least waitFrom but at most waitTo ms
func wait(waitFrom int, waitTo int) {
	wait := waitFrom + (rand.Int() % (waitTo - waitFrom))

	time.Sleep(time.Duration(wait) * time.Millisecond)
}

func loadCerts(certs string, hclient *http.Client) error {
	rt := hclient.Transport.(*h2quic.QuicRoundTripper)
	rt.TLSClientConfig.InsecureSkipVerify = false

	// Load CA certs
	caCert, err := ioutil.ReadFile(certs)

	if err != nil {
		return err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	rt.TLSClientConfig.RootCAs = caCertPool

	return nil
}

func main() {
	var urls = flag.String("urls", "", "URLs to fetch (; delimited).")
	var logfilePath = flag.String("logfile", "puic-poll.log", "File to write debug information to.")
	var waitFrom = flag.Int("wait-from", 1000, "Minimum time to wait in milliseconds before making the next request.")
	var waitTo = flag.Int("wait-to", 5000, "Maximum time to wait in milliseconds before making the next request.")
	var collect = flag.Int("collect", 1024, "How many statistics items to collect in a single output file.")
	var odir = flag.String("odir", "./tmp/", "Output directory.")
	var ifaceName = flag.String("iface", "op0", "Interface to use.")
	var certs = flag.String("certs", "", "Path to certificates to be trusted as Root CAs.")

	flag.Parse()

	
	// Logging setup
	collectedStats := make([]Stats, *collect)
	j := 0

	logfile := openLogfile(*logfilePath)
	defer logfile.Close()
	
	ofile, err := openNextOutputFile(*odir)

	if err != nil {
		writeOrDie(logfile, "ERR: Error opening output file: %q", err.Error())
		panic(err)
	}

	hclient, err := createHttpClient(*ifaceName, logfile)

	if err != nil {
		writeOrDie(logfile, "ERR: Error creating h2client: %q", err.Error())
		panic(err)
	}

	if *certs != "" {
		err := loadCerts(*certs, hclient)

		if err != nil {
			writeOrDie(logfile, "ERR: Error loading certs: %q", err.Error())
			panic(err)
		}
	}

	urlsToFetch := strings.Split(*urls, ";")

	for {

		urlToFetch := urlsToFetch[rand.Int() % len(urlsToFetch)]

		size, speed, elapsed, statusCode, err := fetchOnce(urlToFetch, hclient, logfile)

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

			ofile, err = openNextOutputFile(*odir)

			if err != nil {
				writeOrDie(logfile, "ERR: Error opening output file: %q", err.Error())
				panic(err)
			}
		}

		wait(*waitFrom, *waitTo)
	}
}

// Make a single GET request to url using hclient while logging to log
func fetchOnce(url string, hclient *http.Client, log io.Writer) (int64, float64, float64, int, error) {
	writeOrDie(log, "Start fetching %q", url)

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
