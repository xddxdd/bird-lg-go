package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/magiconair/properties/assert"
)

func TestApiServerListHandler(t *testing.T) {
	setting.servers = []string{"alpha", "beta", "gamma"}
	response := apiServerListHandler(apiRequest{})

	assert.Equal(t, len(response.Result), 3)
	assert.Equal(t, response.Result[0].(apiGenericResultPair).Server, "alpha")
	assert.Equal(t, response.Result[1].(apiGenericResultPair).Server, "beta")
	assert.Equal(t, response.Result[2].(apiGenericResultPair).Server, "gamma")
}

func TestApiGenericHandlerFactory(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, BirdSummaryData)
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show protocols"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	request := apiRequest{
		Servers: setting.servers,
		Type:    "bird",
		Args:    "show protocols",
	}

	handler := apiGenericHandlerFactory("bird")
	response := handler(request)

	assert.Equal(t, response.Error, "")

	result := response.Result[0].(*apiGenericResultPair)
	assert.Equal(t, result.Server, "alpha")
	assert.Equal(t, result.Data, BirdSummaryData)
}

func TestApiSummaryHandler(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, BirdSummaryData)
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show protocols"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	request := apiRequest{
		Servers: setting.servers,
		Type:    "summary",
		Args:    "",
	}
	response := apiSummaryHandler(request)

	assert.Equal(t, response.Error, "")

	summary := response.Result[0].(*apiSummaryResultPair)
	assert.Equal(t, summary.Server, "alpha")
	// Protocol list will be sorted
	assert.Equal(t, summary.Data[1].Name, "device1")
	assert.Equal(t, summary.Data[1].Proto, "Device")
	assert.Equal(t, summary.Data[1].Table, "---")
	assert.Equal(t, summary.Data[1].State, "up")
	assert.Equal(t, summary.Data[1].Since, "2021-08-27")
	assert.Equal(t, summary.Data[1].Info, "")
}

func TestApiSummaryHandlerError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock backend error")
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show protocols"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	request := apiRequest{
		Servers: setting.servers,
		Type:    "summary",
		Args:    "",
	}
	response := apiSummaryHandler(request)

	assert.Equal(t, response.Error, "Mock backend error")
}

func TestApiWhoisHandler(t *testing.T) {
	expectedData := "Mock Data"
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response:      expectedData,
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.whoisServer = server.server.Addr().String()

	request := apiRequest{
		Servers: []string{},
		Type:    "",
		Args:    "AS6939",
	}
	response := apiWhoisHandler(request)

	assert.Equal(t, response.Error, "")

	whoisResult := response.Result[0].(apiGenericResultPair)
	assert.Equal(t, whoisResult.Server, "")
	assert.Equal(t, whoisResult.Data, expectedData)
}

func TestApiErrorHandler(t *testing.T) {
	err := errors.New("Mock Error")
	response := apiErrorHandler(err)
	assert.Equal(t, response.Error, "Mock Error")
}

func TestApiHandler(t *testing.T) {
	setting.servers = []string{"alpha", "beta", "gamma"}

	request := apiRequest{
		Servers: []string{},
		Type:    "server_list",
		Args:    "",
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api", bytes.NewReader(requestJson))
	w := httptest.NewRecorder()
	apiHandler(w, r)

	var response apiResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(response.Result), 3)
	// Hard to unmarshal JSON into apiGenericResultPair objects, won't check here
}

func TestApiHandlerBadJSON(t *testing.T) {
	setting.servers = []string{"alpha", "beta", "gamma"}

	r := httptest.NewRequest(http.MethodGet, "/api", strings.NewReader("{bad json}"))
	w := httptest.NewRecorder()
	apiHandler(w, r)

	var response apiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(response.Result), 0)
}

func TestApiHandlerInvalidType(t *testing.T) {
	setting.servers = []string{"alpha", "beta", "gamma"}

	request := apiRequest{
		Servers: setting.servers,
		Type:    "invalid_type",
		Args:    "",
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api", bytes.NewReader(requestJson))
	w := httptest.NewRecorder()
	apiHandler(w, r)

	var response apiResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(response.Result), 0)
}
