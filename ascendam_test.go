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
	ts := getServer(1, 200)
	defer ts.Close()

	client := getClient(time.Millisecond * 50)
	code, err := getCode(ts.URL, client)
	if err != nil {
		t.Error(err)
	}
	if code != 200 {
		t.Error("got unexpected response code:", code)
	}
}

func TestGetCodeTimeoutAwaitingResponseHeaders(t *testing.T) {
	ts := getServer(50, 200)
	defer ts.Close()
	client := getClient(time.Millisecond * 10)

	_, err := getCode(ts.URL, client)

	if err == nil {
		t.Error("Expected a timeout")
	}

	if e, ok := err.(net.Error); ok && !e.Timeout() {
		t.Error("Expected a timeout error not", err)
	}
}

func TestGetCode404(t *testing.T) {
	ts := getServer(1, 404)
	defer ts.Close()

	client := getClient(time.Millisecond * 50)
	code, err := getCode(ts.URL, client)
	if err != nil {
		t.Errorf("Got unexpected error message: '%s'", err)
		return
	}

	if code != 404 {
		t.Errorf("Expected response code 404, got '%d':", code)
	}
}

func TestGetState200Code(t *testing.T) {
	ts := getServer(1, 200)
	defer ts.Close()

	client := getClient(time.Millisecond * 50)
	up, message := getState(ts.URL, client)

	if up != UP {
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
	ts := getServer(1, 404)
	defer ts.Close()

	client := getClient(time.Millisecond * 50)
	up, message := getState(ts.URL, client)

	if up != DOWN {
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
	ts := getServer(50, 200)
	defer ts.Close()

	client := getClient(time.Millisecond * 30)
	up, message := getState(ts.URL, client)

	if up != DOWN {
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

	client := getClient(time.Millisecond * 30)
	up, message := getState(ts.URL, client)

	if up != DOWN {
		t.Error("Expected webserver to be DOWN, not UP")
	}

	if !strings.Contains(message, "Down") {
		t.Errorf("Message should contain Down, got '%s'", message)
	}

	if !strings.Contains(message, "redirect discovered") {
		t.Errorf("Message should contain 'redirect discovered', got '%s'", message)
	}
}

func getServer(timeout int, respCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(respCode)
		fmt.Fprintln(w, "Hello, client")
		time.Sleep(time.Duration(timeout) * time.Millisecond)
	}))
}
