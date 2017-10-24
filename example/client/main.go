package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/url"
	"sync"
	"crypto/tls"
	"fmt"
	"time"
	//"log"
	"math"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/utils"

	"golang.org/x/net/html"
)

import _ "net/http/pprof"



func follow(addr string, buf io.Reader, urls chan string) {
	tokenizer := html.NewTokenizer(buf)

	base, err := url.Parse(addr)

	if err != nil {
		utils.Infof("Error: %s", err.Error())
		return
	}

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			return
		case html.StartTagToken:
			token := tokenizer.Token()

			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						utils.Infof("href := %s", attr.Val)
						ref, err := url.Parse(attr.Val)

						if err != nil {
							utils.Infof("Error: %s", err.Error())
							continue
						}

						abs := base.ResolveReference(ref)

						nurl := abs.String()

						utils.Infof("Queueing %s", nurl)
						select {
							case	urls <- nurl:
							default:
								utils.Infof("TOO MUCH STUFF")
								return
						}
						utils.Infof("Queued %s", nurl)
					}
				}
			}
		}
	}
}

type stats struct {
	max_speed float64
	min_speed float64
	
	speed_sum float64
	size_sum uint64
	n uint32
	mutex *sync.Mutex
}

func crawl(url string, urls chan string, stats *stats) {
	utils.Infof("Crawling %s", url)

	hclient := &http.Client{
		Transport: &h2quic.QuicRoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify:true}},
	}

	start := time.Now()

	rsp, err := hclient.Get(url)

	if err != nil {
		utils.Infof("Error %s while crawling %s!", err.Error(), url)
		return
	}



	utils.Infof("Status code: %d", rsp.StatusCode)

	body := &bytes.Buffer{}
	_, err = io.Copy(body, rsp.Body)
	
	if err != nil {
		utils.Infof("Error %s while crawling %s!", err.Error(), url)
		return
	}

	end := time.Now()
	elapsed := end.Sub(start).Seconds()

	speed := float64(body.Len())/elapsed

	stats.mutex.Lock()

	if stats.max_speed < speed {
		stats.max_speed = speed
	}

	if stats.min_speed > speed {
		stats.min_speed = speed
	}

	stats.speed_sum += speed
	stats.size_sum += uint64(body.Len())
	stats.n += 1

	stats.mutex.Unlock()

	utils.Infof("Speed: %f, elapsed: %f", speed / (1024.0*1024.0), elapsed)

	rsp.Body.Close()

	if rsp.StatusCode == 200 {
		follow(url, body, urls)
	}
}

func crawlWorker(urls chan string, stats *stats) {
	for {
		
	}
}

func crawlLoop(urls chan string, stats *stats, maxOutstanding int) {
	var wg sync.WaitGroup
	outstanding := 0
	started := false
	done := make(chan bool)

	for {
		utils.Infof("Wait for next url... (%d)", maxOutstanding)
		select {
			case url, ok := <- urls:
				started = true

				if ok {
					if outstanding >= maxOutstanding {
						utils.Infof("Too many open requests.... waiting for slot.")
						_ = <- done
						outstanding -= 1
					}

					outstanding += 1
					wg.Add(1)
					go func() {
						crawl(url, urls, stats)
						wg.Done()
						done <- true
					}()
				} else {
					return
				}

			default:
				if !started {
					continue
				}

				if outstanding == 0 {
					utils.Infof("...")
					return
				} else {
					utils.Infof("Waiting for new data...")
					_ = <- done
					outstanding -= 1
				}
		}
	}
}

func main() {
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:9090", nil))
	//}()


	verbose := flag.Bool("v", false, "verbose")
	maxOutstanding := flag.Int("o", 1, "Max outstanding (concurrent) requests")
	flag.Parse()
	urls := flag.Args()

	if *verbose {
		utils.SetLogLevel(utils.LogLevelDebug)
	} else {
		utils.SetLogLevel(utils.LogLevelInfo)
	}

	stats := &stats{}
	stats.max_speed = 0.0
	stats.min_speed = math.MaxFloat64
	stats.mutex = &sync.Mutex{}

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, addr := range urls {
		utils.Infof("GET %s", addr)
		ch := make(chan string, 8192)
		go func() {
			crawlLoop(ch, stats, *maxOutstanding)
			wg.Done()
		}()
		ch <- addr
	}
	wg.Wait()

	fmt.Printf("Done...\n\n")
	fmt.Printf("Stats:\n")
	fmt.Printf("  Links crawled: %d\n", stats.n)
	total_mb := float64(stats.size_sum)/(1024.0*1024.0)
	speed_mb := float64(stats.speed_sum)/(1024.0*1024.0)
	fmt.Printf("  Total MiB downloaded: %f\n", total_mb)
	fmt.Printf("  Average file size (MiB): %f\n", total_mb / float64(stats.n))
	fmt.Printf("  Average download speed (MiB/s): %f\n", speed_mb / float64(stats.n))
	fmt.Printf("  Max. speed (MiB/s): %.6f\n", stats.max_speed / (1024.0*1024.0))
	fmt.Printf("  Min. speed (MiB/s): %.6f\n", stats.min_speed / (1024.0*1024.0))
}
