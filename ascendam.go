package main

import (
	"net"
	"net/http"
	"fmt"
	"os"
	"flag"
	"time"
	"log"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

var timeout_ms = time.Duration(30 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout_ms)
}

func main() {

	// Not as extensive time logging, time is good enough
	log.SetFlags(log.Ltime)

	// Set up command line flags
	var url string
	flag.StringVar(&url, "url", "", "the url to check")

	var max_time_ms int
	flag.IntVar(&max_time_ms, "max-ms", 30000, "max milliseconds")

	// setup usage instructions
	flag.Usage = Usage
	flag.Parse()

	if url == "" {
		flag.Usage()
		os.Exit(1)
	}

	timeout_ms := time.Millisecond * time.Duration(max_time_ms)

	transport := http.Transport{
		Dial: dialTimeout,
	}

	client := http.Client{
		Transport: &transport,
	}

	fmt.Printf("Running uptime check on '%s'\n", url)

	fmt.Printf("Timeout is set to %s\n", timeout_ms)

	server_up := false

	for {
		failed := false

		// fetch page
		start := time.Now()
		resp, err := client.Get(url)
		elapsed := time.Since(start)

		down_message := ""

		if err != nil {
			down_message = fmt.Sprintf("Down\t%s\t%s\t\t%s", "n/a", elapsed, err)
			failed = true
			time.Sleep(50 * time.Millisecond)
		} else {
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				down_message = fmt.Sprintf("Down\t%d\t%s", resp.StatusCode, elapsed)
				failed = true
			}

			if timeout_ms > 0 && elapsed > timeout_ms {
				down_message = fmt.Sprintf("Down\t%d\t%s", resp.StatusCode, elapsed)
				failed = true
			}
		}

		if failed && server_up {
			server_up = false
			log.Print(down_message)
		} else if !failed && !server_up {
			server_up = true
			log.Printf("Up\t%d\t%s", resp.StatusCode, elapsed)
		}
	}

}
