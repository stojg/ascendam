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

const (
	UP   = true
	DOWN = false
)

type event struct {
	Date     time.Time
	LoadTime time.Duration
	State    bool
}

type EventList struct {
	Events   []event
	Outages  int
    UpDuration time.Duration
    DownDuration   time.Duration
	LastTime time.Time
}

func (e *EventList) Add(state bool, loadTime time.Duration) {

	// if this is an UP event add now - lastTime to Uptime
	if state == UP && !e.LastTime.IsZero() {
		e.UpDuration += (time.Now().Sub(e.LastTime))
	}

	if state == DOWN && !e.LastTime.IsZero() {
		e.DownDuration += (time.Now().Sub(e.LastTime))
	}

    e.LastTime = time.Now()

    if !state {
        e.Outages += 1
    }

	e.Events = append(e.Events, event{
		Date:     time.Now(),
		State:    state,
		LoadTime: loadTime,
	})
}

func (e *EventList) AvgLoadTime() time.Duration {
	var avgTime time.Duration
	nonOutages := 0
	for _, event := range e.Events {
		if event.State != UP {
			continue
		}
		avgTime = avgTime + event.LoadTime
		nonOutages += 1
	}
    if avgTime.Nanoseconds() == 0 {
        return 0
    }
	return avgTime / time.Duration(nonOutages)
}

var eventList *EventList

func init() {
    eventList = &EventList{}
}

func main() {

    // setup trapping of SIGINT
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        // sig is a ^C, handle it
        for sig := range c {
            fmt.Printf("caught %s\n\n", sig)
            fmt.Printf("%d outages of %d checks \n", eventList.Outages, len(eventList.Events))
            fmt.Printf("Average loadtime: %s \n", eventList.AvgLoadTime())
            fmt.Printf("Downtime: %s \n", eventList.DownDuration)
            fmt.Printf("Uptime: %s \n", eventList.UpDuration)
            os.Exit(0)
        }
    }()

	// Just log events with the timestamp, skip the date
	log.SetFlags(log.Ltime)

	// Set up commandline flags
	var url string
	flag.StringVar(&url, "url", "", "the url to check")
	var max_time_ms int
	flag.IntVar(&max_time_ms, "max-ms", 30000, "max milliseconds")
	// Usage instructions
	flag.Usage = Usage
	flag.Parse()

	// -url is required
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

	// infinity loop
	for {
		state, message := getState(url, getClient(timeout))
		if last_state != state {
			log.Print(message)
			last_state = state
		}
		// don't slam the server
		time.Sleep(1 * time.Millisecond)
	}
}

// getState takes an url and a *http.Client and will return a UP or DOWN state and
// a pre formatted message
func getState(url string, client *http.Client) (state bool, message string) {
	start := time.Now()
	code, err := getCode(url, client)
	elapsed := time.Since(start)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			eventList.Add(DOWN, elapsed)
			return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out")
		}
		if strings.Contains(err.Error(), "use of closed network connection") {
			eventList.Add(DOWN, elapsed)
			return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, "request timed out (cancelled)")
		}
		eventList.Add(DOWN, elapsed)
		return DOWN, fmt.Sprintf("Down\t%s\t%s\t%s", "n/a", elapsed, err)
	}
	if code != 200 {
		eventList.Add(DOWN, elapsed)
		return DOWN, fmt.Sprintf("Down\t%d\t%s\t%s", code, elapsed, "non 200 response code")
	}
	eventList.Add(UP, elapsed)
	return UP, fmt.Sprintf("Up\t%d\t%s", code, elapsed)
}

// getClient will return a *http.Client that has timeouts set and
// disallows redirections.
func getClient(timeout_ms time.Duration) *http.Client {
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

// getCode is simple wrapper for doing a request and return the http status
// code and eventually an error. If there is an error the http status code
// will be set to 0.
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
