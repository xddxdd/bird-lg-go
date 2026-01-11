package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestTracerouteArgsToString(t *testing.T) {
	result := tracerouteArgsToString("traceroute", []string{
		"-a",
		"-b",
		"-c",
	}, []string{
		"google.com",
	})

	assert.Equal(t, result, "traceroute -a -b -c google.com")
}

func TestTracerouteTryExecuteSuccess(t *testing.T) {
	_, err := tracerouteTryExecute("sh", []string{
		"-c",
	}, []string{
		"true",
	})

	if err != nil {
		t.Error(err)
	}
}

func TestTracerouteTryExecuteFail(t *testing.T) {
	_, err := tracerouteTryExecute("sh", []string{
		"-c",
	}, []string{
		"false",
	})

	if err == nil {
		t.Error("Should trigger error, not triggered")
	}
}

func TestTracerouteDetectSuccess(t *testing.T) {
	result := tracerouteDetect("sh", []string{
		"-c",
		"true",
	})

	assert.Equal(t, result, true)
}

func TestTracerouteDetectFail(t *testing.T) {
	result := tracerouteDetect("sh", []string{
		"-c",
		"false",
	})

	assert.Equal(t, result, false)
}

func TestTracerouteAutodetect(t *testing.T) {
	pathBackup := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", pathBackup)

	setting.tr_bin = ""
	setting.tr_flags = []string{}
	tracerouteAutodetect()
	// Should not panic
}

func TestTracerouteAutodetectExisting(t *testing.T) {
	setting.tr_bin = "mock"
	setting.tr_flags = []string{"mock"}
	tracerouteAutodetect()
	assert.Equal(t, setting.tr_bin, "mock")
	assert.Equal(t, setting.tr_flags, []string{"mock"})
}

func TestTracerouteAutodetectFlagsOnly(t *testing.T) {
	pathBackup := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", pathBackup)

	setting.tr_bin = "mock"
	setting.tr_flags = nil
	tracerouteAutodetect()

	// Should not panic
}

func TestTracerouteHandlerWithoutQuery(t *testing.T) {
	initTracerouteSemaphore(setting.tr_max_concurrent)
	r := httptest.NewRequest(http.MethodGet, "/traceroute", nil)
	w := httptest.NewRecorder()
	tracerouteHandler(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	if !strings.Contains(w.Body.String(), "Invalid Request") {
		t.Error("Did not get invalid request")
	}
}

func TestTracerouteHandlerShlexError(t *testing.T) {
	initTracerouteSemaphore(setting.tr_max_concurrent)
	r := httptest.NewRequest(http.MethodGet, "/traceroute?q="+url.QueryEscape("\"1.1.1.1"), nil)
	w := httptest.NewRecorder()
	tracerouteHandler(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	if !strings.Contains(w.Body.String(), "parse") {
		t.Error("Did not get parsing error message")
	}
}

func TestTracerouteHandlerNoTracerouteFound(t *testing.T) {
	initTracerouteSemaphore(setting.tr_max_concurrent)
	setting.tr_bin = ""
	setting.tr_flags = nil

	r := httptest.NewRequest(http.MethodGet, "/traceroute?q="+url.QueryEscape("1.1.1.1"), nil)
	w := httptest.NewRecorder()
	tracerouteHandler(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	if !strings.Contains(w.Body.String(), "not supported") {
		t.Error("Did not get not supported error message")
	}
}

func TestTracerouteHandlerExecuteError(t *testing.T) {
	initTracerouteSemaphore(setting.tr_max_concurrent)
	setting.tr_bin = "sh"
	setting.tr_flags = []string{"-c", "false"}
	setting.tr_raw = true

	r := httptest.NewRequest(http.MethodGet, "/traceroute?q="+url.QueryEscape("1.1.1.1"), nil)
	w := httptest.NewRecorder()
	tracerouteHandler(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	if !strings.Contains(w.Body.String(), "Error executing traceroute") {
		t.Error("Did not get not execute error message")
	}
}

func TestTracerouteHandlerRaw(t *testing.T) {
	initTracerouteSemaphore(setting.tr_max_concurrent)
	setting.tr_bin = "sh"
	setting.tr_flags = []string{"-c", "echo Mock"}
	setting.tr_raw = true

	r := httptest.NewRequest(http.MethodGet, "/traceroute?q="+url.QueryEscape("1.1.1.1"), nil)
	w := httptest.NewRecorder()
	tracerouteHandler(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Mock\n")
}

func TestTracerouteHandlerPostprocess(t *testing.T) {
	initTracerouteSemaphore(setting.tr_max_concurrent)
	setting.tr_bin = "sh"
	setting.tr_flags = []string{"-c", "echo \"first line\n 2 *\nthird line\""}
	setting.tr_raw = false

	r := httptest.NewRequest(http.MethodGet, "/traceroute?q="+url.QueryEscape("1.1.1.1"), nil)
	w := httptest.NewRecorder()
	tracerouteHandler(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "first line\nthird line\n\n1 hops not responding.")
}

func TestTracerouteHandlerConcurrencyLimit(t *testing.T) {
	// Set a low limit for testing
	maxConcurrent := 2
	initTracerouteSemaphore(maxConcurrent)
	setting.tr_max_concurrent = maxConcurrent

	// Use a slow command to keep requests running
	setting.tr_bin = "sh"
	setting.tr_flags = []string{"-c", "sleep 1; echo Done"}
	setting.tr_raw = true

	// Launch more concurrent requests than the limit
	numRequests := 5
	responses := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			r := httptest.NewRequest(http.MethodGet, "/traceroute?q="+url.QueryEscape("1.1.1.1"), nil)
			w := httptest.NewRecorder()
			tracerouteHandler(w, r)
			responses <- w.Code
		}()
	}

	// Collect all responses
	statusCodes := make(map[int]int)
	for i := 0; i < numRequests; i++ {
		code := <-responses
		statusCodes[code]++
	}

	// Verify that some requests succeeded (200) and some were rejected (503)
	if statusCodes[http.StatusOK] == 0 {
		t.Error("Expected at least one request to succeed with 200")
	}
	if statusCodes[http.StatusServiceUnavailable] == 0 {
		t.Error("Expected at least one request to be rejected with 503")
	}

	// Verify we didn't get any unexpected status codes
	totalRequests := statusCodes[http.StatusOK] + statusCodes[http.StatusServiceUnavailable]
	if totalRequests != numRequests {
		t.Errorf("Expected %d total requests, got %d", numRequests, totalRequests)
	}
}
