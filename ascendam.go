package main

import (
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

	fmt.Printf("Running uptime check on '%s'\n", url)

	fmt.Printf("Timeout is set to %s\n", timeout_ms)

	server_up := true
	for {

		failed := false

		// fetch page
		start := time.Now()
		resp, err := http.Get(url)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Println("Server unavailable")
			os.Exit(2)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 || (timeout_ms > 0 && elapsed > timeout_ms) {
			failed = true
		}

		if failed && server_up {
			server_up = false
			log.Printf("Down\t%d\t%s", resp.StatusCode, elapsed)
		} else if !failed && !server_up {
			server_up = true
			log.Printf("Up\t%d\t%s", resp.StatusCode, elapsed)
		}
	}




}
