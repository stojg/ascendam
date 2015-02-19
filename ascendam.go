package main

import (
	"net/http"
	"fmt"
	"os"
	"flag"
	"log"
	"time"
	"io/ioutil"
	"crypto/tls"
	"github.com/mreiferson/go-httpclient"
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

	var transport = &httpclient.Transport {
		ConnectTimeout: timeout_ms,
		ResponseHeaderTimeout: timeout_ms,
		RequestTimeout: timeout_ms,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var client = &http.Client{
		Transport: transport,
	}

	fmt.Printf("Running uptime check on '%s'\n", url)

	fmt.Printf("Timeout is set to %s\n", timeout_ms)

	last_state := true

	var message string

	for {
		state := false

		// fetch page
		start := time.Now()
		code, err := getCode(url, client)
		elapsed := time.Since(start)

		if err != nil {
			message = fmt.Sprintf("Down\t%s\t%s\t\t%s", "n/a", elapsed, err)
			state = true
			time.Sleep(50 * time.Millisecond)
		} else if code != 200 {
			message = fmt.Sprintf("Down\t%d\t%s", code, elapsed)
			state = true
		} else  if timeout_ms > 0 && elapsed > timeout_ms {
			message = fmt.Sprintf("Down\t%d\t%s", code, elapsed)
			state = true
		} else {
			message = fmt.Sprintf("Up\t%d\t%s", code, elapsed)
		}

		if last_state != state {
			log.Print(message)
			last_state = state
		}
	}
}

func getCode(url string, client *http.Client) (code int, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "MustWin/health-checker")
	resp, err := client.Do(req)

	if err != nil {
		return 0, err
	}
	defer func() {
		resp.Body.Close()
	}()

	_, err = ioutil.ReadAll(resp.Body)
	if (err != nil) {
		return 0, err
	}

	return resp.StatusCode, nil
}
