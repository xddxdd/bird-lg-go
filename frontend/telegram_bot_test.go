package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/magiconair/properties/assert"
)

func doTestTelegramIsCommand(t *testing.T, message string, command string, expected bool) {
	result := telegramIsCommand(message, command)
	assert.Equal(t, result, expected)
}

func mockTelegramCall(t *testing.T, msg string, raw bool) string {
	return mockTelegramEndpointCall(t, "/telegram/", msg, raw)
}

func mockTelegramEndpointCall(t *testing.T, endpoint string, msg string, raw bool) string {
	request := tgWebhookRequest{
		Message: tgMessage{
			MessageID: 123,
			Chat: tgChat{
				ID: 456,
			},
			Text: msg,
		},
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	requestBody := bytes.NewReader(requestJson)

	r := httptest.NewRequest(http.MethodGet, endpoint, requestBody)
	w := httptest.NewRecorder()
	webHandlerTelegramBot(w, r)

	if raw {
		return w.Body.String()
	} else {
		var response tgWebhookResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error(err)
		}

		assert.Equal(t, response.ChatID, request.Message.Chat.ID)
		assert.Equal(t, response.ReplyToMessageID, request.Message.MessageID)
		return response.Text
	}
}

func TestTelegramIsCommand(t *testing.T) {
	setting.telegramBotName = "test_bot"

	// Recognize command
	doTestTelegramIsCommand(t, "/trace", "trace", true)
	doTestTelegramIsCommand(t, "/trace", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace", "tra", false)
	doTestTelegramIsCommand(t, "/trace", "abcdefg", false)

	// Recognize command with parameters
	doTestTelegramIsCommand(t, "/trace google.com", "trace", true)
	doTestTelegramIsCommand(t, "/trace google.com", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace google.com", "tra", false)
	doTestTelegramIsCommand(t, "/trace google.com", "abcdefg", false)

	// Recognize command with bot name
	doTestTelegramIsCommand(t, "/trace@test_bot", "trace", true)
	doTestTelegramIsCommand(t, "/trace@test_bot", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace@test_bot", "tra", false)
	doTestTelegramIsCommand(t, "/trace@test_bot", "abcdefg", false)
	doTestTelegramIsCommand(t, "/trace@test_bot_123", "trace", false)
	doTestTelegramIsCommand(t, "/trace@test_", "trace", false)

	// Recognize command with bot name and parameters
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "trace", true)
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "tra", false)
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "abcdefg", false)
	doTestTelegramIsCommand(t, "/trace@test_bot_123 google.com", "trace", false)
	doTestTelegramIsCommand(t, "/trace@test google.com", "trace", false)
}

func TestTelegramBatchRequestFormatSingleServer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock")
	httpmock.RegisterResponder("GET", "http://alpha:8000/mock?q=cmd", httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	result := telegramBatchRequestFormat(setting.servers, "mock", "cmd", telegramDefaultPostProcess)
	expected := "Mock\n\n"
	assert.Equal(t, result, expected)
}

func TestTelegramBatchRequestFormatMultipleServers(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock")
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

	result := telegramBatchRequestFormat(setting.servers, "mock", "cmd", telegramDefaultPostProcess)
	expected := "alpha\nMock\n\nbeta\nMock\n\ngamma\nMock\n\n"
	assert.Equal(t, result, expected)
}

func TestWebHandlerTelegramBotBadJSON(t *testing.T) {
	requestBody := strings.NewReader("{bad json}")

	r := httptest.NewRequest(http.MethodGet, "/telegram/", requestBody)
	w := httptest.NewRecorder()
	webHandlerTelegramBot(w, r)

	response := w.Body.String()
	assert.Equal(t, response, "")
}

func TestWebHandlerTelegramBotTrace(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock Response")
	httpmock.RegisterResponder("GET", "http://alpha:8000/traceroute?q=1.1.1.1", httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	response := mockTelegramCall(t, "/trace 1.1.1.1", false)
	assert.Equal(t, response, "```\nMock Response\n```")
}

func TestWebHandlerTelegramBotTraceWithServerList(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock Response")
	httpmock.RegisterResponder("GET", "http://alpha:8000/traceroute?q=1.1.1.1", httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	response := mockTelegramEndpointCall(t, "/telegram/alpha", "/trace 1.1.1.1", false)
	assert.Equal(t, response, "```\nMock Response\n```")
}

func TestWebHandlerTelegramBotRoute(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "Mock Response")
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show route for 1.1.1.1 primary"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	response := mockTelegramCall(t, "/route 1.1.1.1", false)
	assert.Equal(t, response, "```\nMock Response\n```")
}

func TestWebHandlerTelegramBotPath(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, `
BGP.as_path: 123 456
`)
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show route for 1.1.1.1 all primary"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	response := mockTelegramCall(t, "/path 1.1.1.1", false)
	assert.Equal(t, response, "```\n123 456\n```")
}

func TestWebHandlerTelegramBotPathMissing(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpResponse := httpmock.NewStringResponder(200, "No path in this response")
	httpmock.RegisterResponder("GET", "http://alpha:8000/bird?q="+url.QueryEscape("show route for 1.1.1.1 all primary"), httpResponse)

	setting.servers = []string{"alpha"}
	setting.domain = ""
	setting.proxyPort = 8000

	response := mockTelegramCall(t, "/path 1.1.1.1", false)
	assert.Equal(t, response, "```\nempty result\n```")
}

func TestWebHandlerTelegramBotWhois(t *testing.T) {
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

	response := mockTelegramCall(t, "/whois AS6939", false)
	assert.Equal(t, response, "```"+server.response+"```")
}

func TestWebHandlerTelegramBotWhoisDN42Mode(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS4242422547",
		response: `
Query for AS4242422547
`,
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.netSpecificMode = "dn42"
	setting.whoisServer = server.server.Addr().String()

	response := mockTelegramCall(t, "/whois 2547", false)
	assert.Equal(t, response, "```"+server.response+"```")
}

func TestWebHandlerTelegramBotWhoisDN42ModeFullASN(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS4242422547",
		response: `
Query for AS4242422547
`,
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.netSpecificMode = "dn42"
	setting.whoisServer = server.server.Addr().String()

	response := mockTelegramCall(t, "/whois 4242422547", false)
	assert.Equal(t, response, "```"+server.response+"```")
}

func TestWebHandlerTelegramBotWhoisShortenMode(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response: `
Information line that will be removed

# Comment that will be removed
Name: Redacted for privacy
Descr: This is a vvvvvvvvvvvvvvvvvvvvvvveeeeeeeeeeeeeeeeeeeerrrrrrrrrrrrrrrrrrrrrrrryyyyyyyyyyyyyyyyyyy long line that will be skipped.
Looooooooooooooooooooooong key: this line will be skipped.

Preserved1: this line isn't removed.
Preserved2: this line isn't removed.
Preserved3: this line isn't removed.
Preserved4: this line isn't removed.
Preserved5: this line isn't removed.

`,
	}

	expectedResult := `Preserved1: this line isn't removed.
Preserved2: this line isn't removed.
Preserved3: this line isn't removed.
Preserved4: this line isn't removed.
Preserved5: this line isn't removed.

3 line(s) skipped.`

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.netSpecificMode = "shorten"
	setting.whoisServer = server.server.Addr().String()

	response := mockTelegramCall(t, "/whois AS6939", false)
	assert.Equal(t, response, "```\n"+expectedResult+"\n```")
}

func TestWebHandlerTelegramBotHelp(t *testing.T) {
	response := mockTelegramCall(t, "/help", false)
	if !strings.Contains(response, "/trace") {
		t.Error("Did not get help message")
	}
}

func TestWebHandlerTelegramBotUnknownCommand(t *testing.T) {
	response := mockTelegramCall(t, "/nonexistent", true)
	assert.Equal(t, response, "")
}

func TestWebHandlerTelegramBotNotCommand(t *testing.T) {
	response := mockTelegramCall(t, "random chat message", true)
	assert.Equal(t, response, "")
}

func TestWebHandlerTelegramBotEmptyResponse(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response:      "",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.netSpecificMode = ""
	setting.whoisServer = server.server.Addr().String()

	response := mockTelegramCall(t, "/whois AS6939", false)
	assert.Equal(t, response, "```\nempty result\n```")
}

func TestWebHandlerTelegramBotTruncateLongResponse(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response:      strings.Repeat("A", 65536),
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.netSpecificMode = ""
	setting.whoisServer = server.server.Addr().String()

	response := mockTelegramCall(t, "/whois AS6939", false)
	assert.Equal(t, response, "```\n"+strings.Repeat("A", 4096)+"\n```")
}
