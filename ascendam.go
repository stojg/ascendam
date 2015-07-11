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
	eventList    *EventList
	url          string
	timeoutSec   int
	sleepTimeSec int
)

const (
	UP   = true
	DOWN = false
)

func init() {
	eventList = &EventList{}

	flag.StringVar(&url, "url", "", "the url to monitor")
	flag.IntVar(&timeoutSec, "timeout", 30, "in seconds")
	flag.IntVar(&sleepTimeSec, "sleep", 1, "Time between checks, in seconds")
	flag.Usage = Usage
	flag.Parse()
}

func main() {

	validateFlags()

	// setup trapping of SIGINT
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		// handle signal interrupts
		for range sig {
			fmt.Printf("%d outages of %d checks \n", eventList.Outages, len(eventList.Events))
			fmt.Printf("Average loadtime: %s \n", eventList.AvgLoadTime())
			fmt.Printf("Downtime: %s \n", eventList.DownDuration)
			fmt.Printf("Uptime: %s \n", eventList.UpDuration)
			os.Exit(0)
		}
	}()

	timeout := time.Second * time.Duration(timeoutSec)

	fmt.Printf("Running uptime check on '%s'\n", url)
	fmt.Printf("Timeout is set to %s\n", timeout)

	// initial run, always show state
	lastState, message := checkURL(url, getHTTPClient(timeout))
	log.Print(message)

	for {
		state, message := checkURL(url, getHTTPClient(timeout))
		if lastState != state {
			log.Print(message)
			lastState = state
		}
		time.Sleep(time.Duration(sleepTimeSec) * time.Second)
	}
}

func validateFlags() {
	// url is required
	if url == "" {
		flag.Usage()
		os.Exit(1)
	}
}

// checkURL takes an url and a *http.Client and will return a UP or DOWN state and
// a pre formatted message
func checkURL(url string, client *http.Client) (state bool, message string) {
	start := time.Now()
	statusCode, err := getHTTPStatus(url, client)
	elapsed := time.Since(start)

	if err == nil {
		if statusCode != 200 {
			eventList.Add(DOWN, elapsed)
			return DOWN, fmt.Sprintf("Down\t%d\t%s\t%s", statusCode, elapsed, "non 200 response code")
		}
		eventList.Add(UP, elapsed)
		return UP, fmt.Sprintf("Up\t%d\t%s", statusCode, elapsed)
	}

	// this is a network timeout error
	if e, ok := err.(net.Error); ok && e.Timeout() {
		eventList.Add(DOWN, elapsed)
		return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out")
	}

	if strings.Contains(err.Error(), "use of closed network connection") {
		eventList.Add(DOWN, elapsed)
		return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out (cancelled)")
	}

	if strings.Contains(err.Error(), "request canceled while waiting for connection") {
		eventList.Add(DOWN, elapsed)
		return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out (can't connect)")
	}

	if strings.Contains(err.Error(), "connection reset by peer") {
		eventList.Add(DOWN, elapsed)
		return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "connection denied (reset by peer)")
	}

	eventList.Add(DOWN, elapsed)
	return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, err)

}

// getHTTPClient will return a *http.Client with a connection timeout and
// disallows redirections.
func getHTTPClient(timeout_ms time.Duration) *http.Client {
	transport := &httpclient.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: false,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout_ms,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("redirect discovered")
		},
	}
}

// getHTTPStatus is simple wrapper for doing a request and return the http status
// code and eventually an error. If there is an error the http status code
// will be set to 0.
func getHTTPStatus(url string, client *http.Client) (code int, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "github.com/stojg/ascendam")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return 0, err
	}

	if len(body) < 1 {
		return code, errors.New("Response body is 0 bytes")
	}

	return resp.StatusCode, nil
}
