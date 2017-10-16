package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/url"
	"sync"
	"crypto/tls"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/utils"

	"golang.org/x/net/html"
)

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
						urls <- nurl
						utils.Infof("Queued %s", nurl)
					}
				}
			}
		}
	}
}

func crawl(url string, urls chan string) {
	utils.Infof("Crawling %s", url)

	hclient := &http.Client{
		Transport: &h2quic.QuicRoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify:true}},
	}

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

	rsp.Body.Close()

	if rsp.StatusCode == 200 {
		follow(url, body, urls)
	}
}

func crawlJob(jobs chan string, urls chan string) {
	for {
		select {
		case url, ok := <- jobs:
			if ok {
				utils.Infof("Going to crawl %s...", url)
				crawl(url, urls)
			} else {
				return
			}
		}
	}
}

func crawlLoop(urls chan string) {
	jobs := make(chan string)

	go crawlJob(jobs, urls)
	go crawlJob(jobs, urls)
	go crawlJob(jobs, urls)

	for {
		utils.Infof("Wait for next url...")
		select {
			case url, ok := <- urls:
				if ok {
					utils.Infof("Adding %s to crawl list...", url)
					go func() {
						jobs <- url
					}()
				} else {
					return
				}
		}
	}
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()
	urls := flag.Args()

	if *verbose {
		utils.SetLogLevel(utils.LogLevelDebug)
	} else {
		utils.SetLogLevel(utils.LogLevelInfo)
	}

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, addr := range urls {
		utils.Infof("GET %s", addr)
		ch := make(chan string)
		go func() {
			crawlLoop(ch)
			wg.Done()
		}()
		ch <- addr
	}
	wg.Wait()
}
