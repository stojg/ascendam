package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetCodeServerUp(t *testing.T) {
	server := getTestServer(1, 200)
	defer server.Close()

	client := getHTTPClient(time.Millisecond * 50)
	check := doCheck(server.URL, client)
	if check.Error() != nil {
		t.Error(check.Error())
	}
	if check.StatusCode() != 200 {
		t.Errorf("Expected response code 200, got '%d':", check.StatusCode())
	}
}

func TestGetCodeTimeoutAwaitingResponseHeaders(t *testing.T) {
	ts := getTestServer(50, 200)
	defer ts.Close()

	check := doCheck(ts.URL, getHTTPClient(time.Millisecond*10))

	err := check.Error()
	if err == nil {
		t.Error("Expected a timeout")
	}

	if e, ok := err.(net.Error); ok && !e.Timeout() {
		t.Error("Expected a timeout error not", err)
	}
}

func TestGetCode404(t *testing.T) {
	ts := getTestServer(1, 404)
	defer ts.Close()

	check := doCheck(ts.URL, getHTTPClient(time.Millisecond*50))

	err := check.Error()
	if err != nil {
		t.Errorf("Got unexpected error message: '%s'", err)
		return
	}

	if check.StatusCode() != 404 {
		t.Errorf("Expected response code 404, got '%d':", check.StatusCode())
	}
}

func TestGetState200Code(t *testing.T) {
	ts := getTestServer(1, 200)
	defer ts.Close()

	check := checkURL(ts.URL, getHTTPClient(time.Millisecond*50))
	message := getResult(check)

	if !check.Ok() {
		t.Error("Expected webserver to be UP, not DOWN")
	}

	if !strings.Contains(message, "200") {
		t.Errorf("Message should contain 404, got '%s'", message)
	}

	if !strings.Contains(message, "Up") {
		t.Errorf("Message should contain Up, got '%s'", message)
	}
}

func TestGetStateNon200Code(t *testing.T) {
	ts := getTestServer(1, 404)
	defer ts.Close()

	check := checkURL(ts.URL, getHTTPClient(time.Millisecond*50))
	message := getResult(check)

	if !check.Ok() {
		t.Error("Expected webserver to be DOWN, not UP")
	}

	if !strings.Contains(message, "404") {
		t.Errorf("Message should contain 404, got '%s'", message)
	}

	if !strings.Contains(message, "Down") {
		t.Errorf("Message should contain Down, got '%s'", message)
	}
}

func TestGetStateTimeout(t *testing.T) {
	ts := getTestServer(50, 200)
	defer ts.Close()

	check := checkURL(ts.URL, getHTTPClient(time.Millisecond*30))
	message := getResult(check)

	if check.Ok() {
		t.Error("Expected webserver to be DOWN, not UP")
	}

	if !strings.Contains(message, "Down") {
		t.Errorf("Message should contain Down, got '%s'", message)
	}

	if !strings.Contains(message, "timed out") {
		t.Errorf("Message should contain 'timeout', got '%s'", message)
	}
}

func TestGetStateRedirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/yolo", 302)
	}))
	defer ts.Close()

	check := checkURL(ts.URL, getHTTPClient(time.Millisecond*30))
	message := getResult(check)

	if check.Ok() {
		t.Error("Expected webserver to be DOWN, not UP")
	}

	if !strings.Contains(message, "Down") {
		t.Errorf("Message should contain Down, got '%s'", message)
	}

	if !strings.Contains(message, "redirect discovered") {
		t.Errorf("Message should contain 'redirect discovered', got '%s'", message)
	}
}

func getTestServer(timeout int, respCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(respCode)
		fmt.Fprintln(w, "Hello, client")
		time.Sleep(time.Duration(timeout) * time.Millisecond)
	}))
}
