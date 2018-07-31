package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mreiferson/go-httpclient"
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
	verboseMode  bool
	debugMode    bool
)

func init() {
	checks = &CheckList{}

	flag.StringVar(&url, "url", "", "the url to monitor")
	flag.IntVar(&timeoutSec, "timeout", 4, "in seconds")
	flag.IntVar(&sleepTimeSec, "sleep", 1, "Time between checks, in seconds")
	flag.BoolVar(&verboseMode, "verbose", false, "Be more verbose")
	flag.BoolVar(&debugMode, "debug", false, "Be super verbose")
	flag.Usage = Usage
	flag.Parse()

	if debugMode {
		verboseMode = true
	}
}

func main() {

	validateArguments()

	// setup trapping of SIGINT
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	timeout := time.Second * time.Duration(timeoutSec)
	sleep := time.Second * time.Duration(sleepTimeSec)

	fmt.Printf("Running uptime check on '%s'\n", url)
	fmt.Printf("Timeout is set to %s and pause %s between checks\n", timeout, sleep)
	fmt.Println("Stop ascendam with ctrl+c")

	// Always log the first check
	check := doCheck(url, getHTTPClient(timeout))
	checks.Add(check)
	log.Printf("%s\n", getResult(check))

	ticker := time.NewTicker(sleep)

loop:
	for {
		select {

		case <-ticker.C:
			check := doCheck(url, getHTTPClient(timeout))
			if verboseMode || check.Ok() != checks.lastState {
				log.Printf("%s\n", getResult(check))
			}
			checks.Add(check)
		case <-sig:
			break loop
		}
	}
	printSummary(checks)
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
		DisableKeepAlives: true,
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
	req.Close = true
	if err != nil {
		check.error = err
		return check
	}
	req.Header.Add("User-Agent", "github.com/stojg/ascendam")

	trace := &httptrace.ClientTrace{

		// GetConn is called before a connection is created or
		// retrieved from an idle pool. The hostPort is the
		// "host:port" of the target or proxy. GetConn is called even
		// if there's already an idle cached connection available.
		GetConn: func(hostPort string) {
			debug("Getting connection to %s\n", hostPort)
		},

		// TLSHandshakeStart is called when the TLS handshake is started. When
		// connecting to a HTTPS site via a HTTP proxy, the handshake happens after
		// the CONNECT request is processed by the proxy.
		TLSHandshakeStart: func() {
			debug("Starting TLS handshake negotiation\n")
		},

		// TLSHandshakeDone is called after the TLS handshake with either the
		// successful handshake's connection state, or a non-nil error on handshake
		// failure.
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				debug("Error during TLS handshake negotiation: %s\n", err)
			} else {
				debug("TLS handshake negotiation done")
			}
		},

		GotConn: func(connInfo httptrace.GotConnInfo) {
			debug("Got connection to %s\n", connInfo.Conn.RemoteAddr())
		},

		// GotFirstResponseByte is called when the first byte of the response
		// headers is available.
		GotFirstResponseByte: func() {
			debug("Got first byte\n")
		},

		// Got100Continue is called if the server replies with a "100
		// Continue" response.
		Got100Continue: func() {
			debug("Got100Continue\n")
		},

		// DNSStart is called when a DNS lookup begins.
		DNSStart: func(info httptrace.DNSStartInfo) {
			debug("DNSStart\n")
		},

		// DNSDone is called when a DNS lookup ends.
		DNSDone: func(info httptrace.DNSDoneInfo) {
			debug("DNSDone\n")
		},

		// ConnectStart is called when a new connection's Dial begins.
		// If net.Dialer.DualStack (IPv6 "Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectStart: func(network, addr string) {
			debug("ConnectStart\n")
		},

		// ConnectDone is called when a new connection's Dial
		// completes. The provided err indicates whether the
		// connection completedly successfully.
		// If net.Dialer.DualStack ("Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectDone: func(network, addr string, err error) {
			debug("ConnectDone\n")
		},

		// WroteHeaders is called after the Transport has written
		// the request headers.
		WroteHeaders: func() {
			debug("Wrote the HTTP headers")
		},

		// Wait100Continue is called if the Request specified
		// "Expected: 100-continue" and the Transport has written the
		// request headers but is waiting for "100 Continue" from the
		// server before writing the request body.
		Wait100Continue: func() {
			debug("Wait100Continue")
		},

		// WroteRequest is called with the result of writing the
		// request and any body. It may be called multiple times
		// in the case of retried requests.
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			debug("Wrote the request with body")
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

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
		check.error = errors.New("response body is 0 bytes")
		return check
	}

	check.statusCode = resp.StatusCode
	return check
}

func debug(format string, v ...interface{}) {
	if debugMode {
		log.Printf(format, v...)
	}
}
