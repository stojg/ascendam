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
	"os/signal"
	"strings"
	"time"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of ascendam:\n")
	flag.PrintDefaults()
}

var (
	checks       *CheckList
	url          string
	timeoutSec   int
	sleepTimeSec int
	verbose      bool
)

func init() {
	checks = &CheckList{}

	flag.StringVar(&url, "url", "", "the url to monitor")
	flag.IntVar(&timeoutSec, "timeout", 4, "in seconds")
	flag.IntVar(&sleepTimeSec, "sleep", 1, "Time between checks, in seconds")
	flag.BoolVar(&verbose, "verbose", false, "Be more verbose")
	flag.Usage = Usage
	flag.Parse()
}

func main() {

	validateArguments()

	// setup trapping of SIGINT
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		// handle signal interrupts
		for _ = range sig {
			printSummary(checks)
			os.Exit(0)
		}
	}()

	timeout := time.Second * time.Duration(timeoutSec)
	sleep := time.Second * time.Duration(sleepTimeSec)

	fmt.Printf("Running uptime check on '%s'\n", url)
	fmt.Printf("Timeout is set to %s and pause %s between checks\n", timeout, sleep)
	fmt.Printf("Stop ascendam with ctrl+c\n")

	// Always log the first check
	check := doCheck(url, getHTTPClient(timeout))
	checks.Add(check)
	log.Printf("%s\n", getResult(check))

	// @todo(stig): emit a summary every minute
	for {
		select {
		case <-time.After(time.Minute):
			printSummary(checks)

		case <-time.After(sleep):
			check := doCheck(url, getHTTPClient(timeout))
			if verbose || check.Ok() != checks.lastState {
				log.Printf("%s\n", getResult(check))
			}
			checks.Add(check)
		}
	}
}

func printSummary(checks *CheckList) {
	fmt.Printf("\n%d outages of %d checks \n", checks.Down(), checks.Total())
	fmt.Printf("Average loadtime: %s \n", checks.AvgLoadTime())
	fmt.Printf("Downtime: %s \n", checks.Downtime())
	fmt.Printf("Uptime: %s \n", checks.Uptime())
}

func validateArguments() {
	// url is required
	if url == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func getResult(check *Check) string {
	elapsed := check.LoadTime()

	if check.Ok() {
		if check.StatusCode() != 200 {
			return fmt.Sprintf("Down\t%d\t%s\t%s", check.StatusCode(), elapsed, "non 200 response code")
		}
		return fmt.Sprintf("Up\t%d\t%s", check.StatusCode(), elapsed)
	}

	err := check.Error()
	// this is a network timeout error
	if e, ok := err.(net.Error); ok && e.Timeout() {
		return fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out")
	}

	if strings.Contains(err.Error(), "use of closed network connection") {
		return fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out (timeout)")
	}

	if strings.Contains(err.Error(), "request canceled while waiting for connection") {
		return fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out (can't connect)")
	}

	if strings.Contains(err.Error(), "connection reset by peer") {
		return fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "connection denied (reset by peer)")
	}

	return fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, err)
}

// getHTTPClient will return a *http.Client with a connection timeout and
// disallows redirections.
func getHTTPClient(timeout time.Duration) *http.Client {
	transport := &httpclient.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: false,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("redirect discovered")
		},
	}
}

// doCheck is simple wrapper for doing a request and return the http status
// code and eventually an error. If there is an error the http status code
// will be set to 0.
func doCheck(url string, client *http.Client) *Check {
	check := &Check{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		check.error = err
		return check
	}
	req.Header.Add("User-Agent", "github.com/stojg/ascendam")

	check.Start()
	resp, err := client.Do(req)
	check.Stop()

	if err != nil {
		check.error = err
		return check
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		check.error = err
		return check
	}

	if len(body) < 1 {
		check.error = errors.New("Response body is 0 bytes")
		return check
	}

	check.statusCode = resp.StatusCode
	return check

}
