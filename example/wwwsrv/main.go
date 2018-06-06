package main

import (
	"flag"
	"net/http"
	"math/rand"

	_ "net/http/pprof"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/utils"
)

func initHttp(prefix string) {
	http.HandleFunc(prefix + "256", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 256)
		rand.Read(data)
		w.Write(data)
	})

	
	http.HandleFunc(prefix + "4KiB", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 4096)
		rand.Read(data)
		w.Write(data)
	})


	http.HandleFunc(prefix + "1MiB", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 1024 * 1024)
		rand.Read(data)
		w.Write(data)
	})

	
	http.HandleFunc(prefix + "16MiB", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 16 * 1024 * 1024)
		rand.Read(data)
		w.Write(data)
	})
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	laddr := flag.String("laddr", "localhost:6121", "Local address to listen on.")
	certPath := flag.String("certpath", "./", "certificate directory. Must contain two files fullchain.pem and privkey.pem.")
	prefix := flag.String("prefix","/data/","Path prefix where the API methods should be available under (should start and end with a slash)")
	flag.Parse()

	initHttp(*prefix)

	if *verbose {
		utils.SetLogLevel(utils.LogLevelDebug)
	} else {
		utils.SetLogLevel(utils.LogLevelInfo)
	}

	certFile := *certPath + "/fullchain.pem"
	keyFile := *certPath + "/privkey.pem"

	err := h2quic.ListenAndServeQUIC(*laddr, certFile, keyFile, nil)

	if err != nil {
		panic(err)
	}
}
