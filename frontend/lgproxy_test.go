package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestBatchRequestIPv4(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock Result")
	httpmock.RegisterResponder("GET", "http://1.1.1.1:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://2.2.2.2:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://3.3.3.3:8000/mock?q=cmd", httpResponse)

	setting.servers = []string{
		"1.1.1.1",
		"2.2.2.2",
		"3.3.3.3",
	}
	setting.domain = ""
	setting.proxyPort = 8000
	response := batchRequest(setting.servers, "mock", "cmd")

	if len(response) != 3 {
		t.Error("Did not get response of all three mock servers")
	}
	for i := 0; i < len(response); i++ {
		if response[i] != "Mock Result" {
			t.Error("HTTP response mismatch")
		}
	}
}

func TestBatchRequestIPv6(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock Result")
	httpmock.RegisterResponder("GET", "http://[2001:db8::1]:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://[2001:db8::2]:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://[2001:db8::3]:8000/mock?q=cmd", httpResponse)

	setting.servers = []string{
		"2001:db8::1",
		"2001:db8::2",
		"2001:db8::3",
	}
	setting.domain = ""
	setting.proxyPort = 8000
	response := batchRequest(setting.servers, "mock", "cmd")

	if len(response) != 3 {
		t.Error("Did not get response of all three mock servers")
	}
	for i := 0; i < len(response); i++ {
		if response[i] != "Mock Result" {
			t.Error("HTTP response mismatch")
		}
	}
}

func TestBatchRequestEmptyResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "")
	httpmock.RegisterResponder("GET", "http://alpha:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://beta:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://gamma:8000/mock?q=cmd", httpResponse)

	setting.servers = []string{
		"alpha",
		"beta",
		"gamma",
	}
	setting.domain = ""
	setting.proxyPort = 8000
	response := batchRequest(setting.servers, "mock", "cmd")

	if len(response) != 3 {
		t.Error("Did not get response of all three mock servers")
	}
	for i := 0; i < len(response); i++ {
		if !strings.Contains(response[i], "node returned empty response") {
			t.Error("Did not produce error for empty response")
		}
	}
}

func TestBatchRequestDomainSuffix(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock Result")
	httpmock.RegisterResponder("GET", "http://alpha.suffix:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://beta.suffix:8000/mock?q=cmd", httpResponse)
	httpmock.RegisterResponder("GET", "http://gamma.suffix:8000/mock?q=cmd", httpResponse)

	setting.servers = []string{
		"alpha",
		"beta",
		"gamma",
	}
	setting.domain = "suffix"
	setting.proxyPort = 8000
	response := batchRequest(setting.servers, "mock", "cmd")

	if len(response) != 3 {
		t.Error("Did not get response of all three mock servers")
	}
	for i := 0; i < len(response); i++ {
		if response[i] != "Mock Result" {
			t.Error("HTTP response mismatch")
		}
	}
}

func TestBatchRequestHTTPError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpError := httpmock.NewErrorResponder(errors.New("Oops!"))
	httpmock.RegisterResponder("GET", "http://alpha:8000/mock?q=cmd", httpError)
	httpmock.RegisterResponder("GET", "http://beta:8000/mock?q=cmd", httpError)
	httpmock.RegisterResponder("GET", "http://gamma:8000/mock?q=cmd", httpError)

	setting.servers = []string{
		"alpha",
		"beta",
		"gamma",
	}
	setting.domain = ""
	setting.proxyPort = 8000
	response := batchRequest(setting.servers, "mock", "cmd")

	if len(response) != 3 {
		t.Error("Did not get response of all three mock servers")
	}
	for i := 0; i < len(response); i++ {
		if !strings.Contains(response[i], "request failed") {
			t.Error("Did not produce HTTP error")
		}
	}
}

func TestBatchRequestInvalidServer(t *testing.T) {
	setting.servers = []string{}
	setting.domain = ""
	setting.proxyPort = 8000
	response := batchRequest([]string{"invalid"}, "mock", "cmd")

	if len(response) != 1 {
		t.Error("Did not get response of all mock servers")
	}
	if !strings.Contains(response[0], "invalid server") {
		t.Error("Did not produce invalid server error")
	}
}
