package api

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"
)

// statusHandler is an http.Handler that writes an empty response using itself
// as the response status code.
type statusHandler int

func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(int(*h))
}

// Check.
func TestIsTagged(t *testing.T) {
    // Set up a fake "Google Code" web server reporting 404 not found.
    status := statusHandler(http.StatusNotFound)
    s := httptest.NewServer(&status)
    defer s.Close()

    //if isTagged(s.URL) {
    //    t.Fatal("isTagged == true, want false")
    //}

    // Change fake server status to 200 OK and try again.
    //status = http.StatusOK

    //if !isTagged(s.URL) {
    //    t.Fatal("isTagged == false, want true")
    //}
}

func TestIntegration(t *testing.T) {
    status := statusHandler(http.StatusNotFound)
    ts := httptest.NewServer(&status)
    defer ts.Close()

    // Replace the pollSleep with a closure that we can block and unblock.
    sleep := make(chan bool)
    pollSleep = func(time.Duration) {
        sleep <- true
        sleep <- true
    }

    // Replace pollDone with a closure that will tell us when the poller is
    // exiting.
    done := make(chan bool)
    pollDone = func() {
        done <- true
    }

    // Put things as they were when the test finishes.
    defer func() {
        pollSleep = time.Sleep
        pollDone = func() {}
    }()

    s := NewServer(ts.URL, 1*time.Millisecond)

    <-sleep // Wait for poll loop to start sleeping.

    // Make first request to the server.
    r, _ := http.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, r)
    if b := w.Body.String(); !strings.Contains(b, "Rate limited") {
        t.Fatalf("body = %s, want Rate limited", b)
    }

    status = http.StatusOK

    <-sleep // Permit poll loop to stop sleeping.
    <-done  // Wait for poller to see the "OK" status and exit.

    // Make second request to the server.
    w = httptest.NewRecorder()
    s.ServeHTTP(w, r)
    if b := w.Body.String(); !strings.Contains(b, "Ping!") {
        t.Fatalf("body = %q, want Ping!", b)
    }
}
