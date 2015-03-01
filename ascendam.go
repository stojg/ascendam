package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/mreiferson/go-httpclient"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of ascendam:\n")
	flag.PrintDefaults()
}

const (
	UP   = true
	DOWN = false
)

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

	timeout := time.Millisecond * time.Duration(max_time_ms)

	fmt.Printf("Running uptime check on '%s'\n", url)

	fmt.Printf("Timeout is set to %s\n", timeout)

	// initial run, always show state
	last_state, message := getState(url, getClient(timeout))
	log.Print(message)

	for {
		state, message := getState(url, getClient(timeout))
		if last_state != state {
			log.Print(message)
			last_state = state
		}
		// don't slam the server
		time.Sleep(100 * time.Millisecond)
	}
}

func getState(url string, client *http.Client) (state bool, message string) {
	start := time.Now()
	code, err := getCode(url, client)
	elapsed := time.Since(start)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out")
		}
		if strings.Contains(err.Error(), "use of closed network connection") {
			return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out (cancelled)")
		}
		return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, err)
	}
	if code != 200 {
		return DOWN, fmt.Sprintf("Down\t%d\t%s\t%s", code, elapsed, "non 200 response code")
	}
	return UP, fmt.Sprintf("Up\t%d\t%s", code, elapsed)
}

func noRedirect(*http.Request, []*http.Request) error {
	return errors.New("redirect discovered")
}

func getClient(timeout_ms time.Duration) *http.Client {
	transport := &httpclient.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: false,
	}
	return &http.Client{
		Transport:     transport,
		Timeout:       timeout_ms,
		CheckRedirect: noRedirect,
	}
}

func getCode(url string, client *http.Client) (code int, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "github.com/stojg/ascendam")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
