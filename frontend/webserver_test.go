package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/magiconair/properties/assert"
)

func TestServerError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	serverError(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
}

func TestWebHandlerWhois(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response:      AS6939Response,
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.netSpecificMode = ""
	setting.whoisServer = server.server.Addr().String()

	r := httptest.NewRequest(http.MethodGet, "/whois/AS6939", nil)
	w := httptest.NewRecorder()
	webHandlerWhois(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	if !strings.Contains(w.Body.String(), "HURRICANE") {
		t.Error("Body does not contain whois result")
	}
}

func TestWebBackendCommunicator(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	input := readDataFile(t, "frontend/test_data/bgpmap_case1.txt")
	httpResponse := httpmock.NewStringResponder(200, input)
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show route for 1.1.1.1 all"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000
	setting.dnsInterface = ""
	setting.whoisServer = ""

	r := httptest.NewRequest(http.MethodGet, "/route_bgpmap/alpha/1.1.1.1", nil)
	w := httptest.NewRecorder()

	handler := webBackendCommunicator("bird", "route_all")
	handler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
}

func TestWebHandlerBGPMap(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	input := readDataFile(t, "frontend/test_data/bgpmap_case1.txt")
	httpResponse := httpmock.NewStringResponder(200, input)
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show route for 1.1.1.1 all"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000
	setting.dnsInterface = ""
	setting.whoisServer = ""

	r := httptest.NewRequest(http.MethodGet, "/route_bgpmap/alpha/1.1.1.1", nil)
	w := httptest.NewRecorder()

	handler := webHandlerBGPMap("bird", "route_bgpmap")
	handler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
}
