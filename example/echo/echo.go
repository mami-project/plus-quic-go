package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"math/big"
	"time"
	"flag"
	"runtime/pprof"
	"runtime"

	quic "github.com/lucas-clemente/quic-go"
   //"github.com/lucas-clemente/quic-go/utils"
	//"plus"
)

const addr = "localhost:4242"

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	mode := flag.Bool("m", true, "Client?")
	usePlus := flag.Bool("plus", true, "PLUS?")
	cpuprofile := flag.String("cp", "", "cpuprofile")
	memprofile := flag.String("mp", "", "memprofile")
	

	flag.Parse()
   //utils.SetLogLevel(utils.LogLevelDebug)
	//PLUS.LoggerDestination = os.Stdout

	if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
	}

	if !(*mode) {
		fmt.Println("SERVER")
		echoServer(*usePlus)
	} else {
		fmt.Println("CLIENT")
		clientMain(*usePlus)
	}

	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal("could not create memory profile: ", err)
        }
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("could not write memory profile: ", err)
        }
        f.Close()
    }
}

// Start a server that echos all data on the first stream opened by the client
func echoServer(usePLUS bool) error {
	cfgServer := &quic.Config{
		TLSConfig: generateTLSConfig(),
        UsePLUS: usePLUS,
	}
	listener, err := quic.ListenAddr(addr, cfgServer)
	if err != nil {
		return err
	}
	sess, err := listener.Accept()
	if err != nil {
		return err
	}
	stream, err := sess.AcceptStream()
	if err != nil {
		panic(err)
	}

	// Copy it all.
	data := make([]byte, 4096)
	bytesRead := uint64(0)

	start_time := time.Now()

	i := 0

	for {
		n, err := stream.Read(data)
		bytesRead += uint64(n)
		fmt.Printf("got something %d %d\n", bytesRead, n)
		if bytesRead >= (1024*1024*50) {
			end_time := time.Now()
			delta := end_time.Sub(start_time).Seconds()
			fmt.Printf("Bytes read so far: %d\n", bytesRead)
			fmt.Printf("Speed: %f MiB/s\n", (float64(bytesRead)/delta)/(1024.0*1024.0))
			bytesRead = 0
			start_time = time.Now()
			i += 1
			if i >= 7 {
				return nil
			}
		}

		stream.Write(data)

		if err != nil {
			return err
		}
	}
}

func clientMain(usePLUS bool) error {
	cfgClient := &quic.Config{
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
        UsePLUS: usePLUS,
	}
	session, err := quic.DialAddr(addr, cfgClient)
	if err != nil {
		return err
	}
    
    fmt.Println("Opening stream")

	stream, err := session.OpenStreamSync()
	if err != nil {
		return err
	}

	fmt.Printf("Client: Sending...\n")

	buf := make([]byte, 4096)
	buf[0] = 1
	buf[1] = 99
	buf[405] = 255

	for i := 0; i < 327680; i++ { 
		_, err = stream.Write(buf)
		if err != nil {
			return err
		}
	}


	return nil
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
