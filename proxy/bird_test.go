package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
)

type BirdServer struct {
	t             *testing.T
	expectedQuery string
	response      string
	server        net.Listener
	socket        string
	injectError   string
}

func (s *BirdServer) initSocket() {
	tmpDir, err := ioutil.TempDir("", "bird-lgproxy-go-mock")
	if err != nil {
		s.t.Fatal(err)
	}
	s.socket = path.Join(tmpDir, "mock.socket")
}

func (s *BirdServer) Listen() {
	s.initSocket()

	var err error
	s.server, err = net.Listen("unix", s.socket)
	if err != nil {
		s.t.Error(err)
	}
}

func (s *BirdServer) Run() {
	for {
		conn, err := s.server.Accept()
		if err != nil {
			break
		}
		if conn == nil {
			break
		}

		reader := bufio.NewReader(conn)

		conn.Write([]byte("1234 Hello from mock bird\n"))

		query, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		if strings.TrimSpace(string(query)) != "restrict" {
			s.t.Errorf("Did not restrict bird permissions")
		}
		if s.injectError == "restriction" {
			conn.Write([]byte("1234 Restriction is disabled!\n"))
		} else {
			conn.Write([]byte("1234 Access restricted\n"))
		}

		query, err = reader.ReadBytes('\n')
		if err != nil {
			break
		}
		if strings.TrimSpace(string(query)) != s.expectedQuery {
			s.t.Errorf("Query %s doesn't match expectation %s", string(query), s.expectedQuery)
		}

		responseList := strings.Split(s.response, "\n")
		for i := range responseList {
			if i == len(responseList)-1 {
				if s.injectError == "eof" {
					conn.Write([]byte("0000 " + responseList[i]))
				} else {
					conn.Write([]byte("0000 " + responseList[i] + "\n"))
				}
			} else {
				conn.Write([]byte("1234 " + responseList[i] + "\n"))
			}
		}

		conn.Close()
	}
}

func (s *BirdServer) Close() {
	if s.server == nil {
		return
	}
	s.server.Close()
}

func TestBirdReadln(t *testing.T) {
	input := strings.NewReader("1234 Bird Message\n")
	var output bytes.Buffer
	birdReadln(input, &output)

	assert.Equal(t, output.String(), "Bird Message\n")
}

func TestBirdReadlnNoPrefix(t *testing.T) {
	input := strings.NewReader(" Message without prefix\n")
	var output bytes.Buffer
	birdReadln(input, &output)

	assert.Equal(t, output.String(), "Message without prefix\n")
}

func TestBirdReadlnVeryLongLine(t *testing.T) {
	input := strings.NewReader(strings.Repeat("A", 4096))
	var output bytes.Buffer
	birdReadln(input, &output)

	assert.Equal(t, output.String(), strings.Repeat("A", 1022)+"\n")
}

func TestBirdWriteln(t *testing.T) {
	var output bytes.Buffer
	birdWriteln(&output, "Test command")
	assert.Equal(t, output.String(), "Test command\n")
}

func TestBirdHandlerWithoutQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/bird", nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)
}

func TestBirdHandlerWithQuery(t *testing.T) {
	server := BirdServer{
		t:             t,
		expectedQuery: "show protocols",
		response:      "Mock Response\nSecond Line",
		injectError:   "",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.birdSocket = server.socket

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape(server.expectedQuery), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), server.response+"\n")
}

func TestBirdHandlerWithBadSocket(t *testing.T) {
	setting.birdSocket = "/nonexistent.sock"

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("mock"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusInternalServerError)
}

func TestBirdHandlerWithoutRestriction(t *testing.T) {
	server := BirdServer{
		t:             t,
		expectedQuery: "show protocols",
		response:      "Mock Response",
		injectError:   "restriction",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.birdSocket = server.socket

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("mock"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusInternalServerError)
}

func TestBirdHandlerEOF(t *testing.T) {
	server := BirdServer{
		t:             t,
		expectedQuery: "show protocols",
		response:      "Mock Response\nSecond Line",
		injectError:   "eof",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.birdSocket = server.socket

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("show protocols"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Mock Response\nEOF")
}

func TestIsBirdCommandAllowed(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"show protocols with params", "show protocols all", true},
		{"show protocols exact match", "show protocols", true},
		{"show protocols with single space", "show protocols ", true},
		{"show route with params", "show route for 1.2.3.4", true},
		{"show route exact match", "show route", true},
		{"show route with single space", "show route ", true},
		{"show status - forbidden", "show status", false},
		{"configure - forbidden", "configure", false},
		{"reload - forbidden", "reload", false},
		{"empty string", "", false},
		{"just show", "show", false},
		{"show protocol - missing s", "show protocol all", false},
		{"show protocolsaaa - no space", "show protocolsaaa", false},
		{"case sensitive", "Show Protocols all", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBirdCommandAllowed(tt.query)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestBirdHandlerWithRestrictedCommandAllowed(t *testing.T) {
	server := BirdServer{
		t:             t,
		expectedQuery: "show protocols all",
		response:      "Mock Response",
		injectError:   "",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.birdSocket = server.socket
	setting.birdRestrictCmds = true

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("show protocols all"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Mock Response\n")
}

func TestBirdHandlerWithRestrictedCommandForbidden(t *testing.T) {
	setting.birdSocket = "/any/path" // Won't be used since we return 403 before connecting
	setting.birdRestrictCmds = true

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("show status"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusForbidden)
	assert.Equal(t, strings.Contains(w.Body.String(), "Forbidden"), true)
}

func TestBirdHandlerWithRestrictedCommandShowRoute(t *testing.T) {
	server := BirdServer{
		t:             t,
		expectedQuery: "show route for 1.2.3.4",
		response:      "Mock Route Response",
		injectError:   "",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.birdSocket = server.socket
	setting.birdRestrictCmds = true

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("show route for 1.2.3.4"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Mock Route Response\n")
}

func TestBirdHandlerWithRestrictedCommandConfigureForbidden(t *testing.T) {
	setting.birdSocket = "/any/path" // Won't be used since we return 403 before connecting
	setting.birdRestrictCmds = true

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("configure"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestBirdHandlerWithRestrictionDisabled(t *testing.T) {
	server := BirdServer{
		t:             t,
		expectedQuery: "show status",
		response:      "Mock Status Response",
		injectError:   "",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.birdSocket = server.socket
	setting.birdRestrictCmds = false

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("show status"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Mock Status Response\n")
}

func TestBirdHandlerWithRestrictionDisabledEdgeCase(t *testing.T) {
	server := BirdServer{
		t:             t,
		expectedQuery: "show protocolsaaa",
		response:      "Mock Response",
		injectError:   "",
	}

	server.Listen()
	go server.Run()
	defer server.Close()

	setting.birdSocket = server.socket
	setting.birdRestrictCmds = false

	r := httptest.NewRequest(http.MethodGet, "/bird?q="+url.QueryEscape("show protocolsaaa"), nil)
	w := httptest.NewRecorder()
	birdHandler(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Mock Response\n")
}
