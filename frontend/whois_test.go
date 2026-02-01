package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
)

func TestAddDefaultWhoisPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// IPv4 addresses
		{"IPv4 without port", "192.0.2.1", "192.0.2.1:43"},
		{"IPv4 with port", "192.0.2.1:4343", "192.0.2.1:4343"},

		// IPv6 addresses - bare format
		{"IPv6 bare without port", "::1", "[::1]:43"},
		{"IPv6 bare full without port", "2001:db8::1", "[2001:db8::1]:43"},

		// IPv6 addresses - bracketed format
		{"IPv6 bracketed without port", "[::1]", "[::1]:43"},
		{"IPv6 bracketed full without port", "[2001:db8::1]", "[2001:db8::1]:43"},
		{"IPv6 bracketed with port", "[::1]:4343", "[::1]:4343"},
		{"IPv6 bracketed full with port", "[2001:db8::1]:4343", "[2001:db8::1]:4343"},

		// Domain names
		{"Domain without port", "whois.example.com", "whois.example.com:43"},
		{"Domain with port", "whois.example.com:4343", "whois.example.com:4343"},

		// Edge cases
		{"Localhost without port", "localhost", "localhost:43"},
		{"Localhost with port", "localhost:4343", "localhost:4343"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addDefaultWhoisPort(tt.input)
			if result != tt.expected {
				t.Errorf("addDefaultWhoisPort(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

type WhoisServer struct {
	t             *testing.T
	expectedQuery string
	response      string
	server        net.Listener
	listenAddr    string // Address to listen on, defaults to "127.0.0.1:0"
}

const AS6939Response = `
ASNumber:       6939
ASName:         HURRICANE
ASHandle:       AS6939
RegDate:        1996-06-28
Updated:        2003-11-04
Ref:            https://rdap.arin.net/registry/autnum/6939
`

func (s *WhoisServer) Listen() error {
	var err error
	listenAddr := s.listenAddr
	if listenAddr == "" {
		listenAddr = "127.0.0.1:0"
	}
	s.server, err = net.Listen("tcp", listenAddr)
	return err
}

func (s *WhoisServer) Run() {
	for {
		conn, err := s.server.Accept()
		if err != nil {
			break
		}
		if conn == nil {
			break
		}

		reader := bufio.NewReader(conn)
		query, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		if strings.TrimSpace(string(query)) != s.expectedQuery {
			s.t.Errorf("Query %s doesn't match expectation %s", string(query), s.expectedQuery)
		}
		conn.Write([]byte(s.response))
		conn.Close()
	}
}

func (s *WhoisServer) Close() {
	if s.server == nil {
		return
	}
	s.server.Close()
}

func TestWhois(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response:      AS6939Response,
	}

	if err := server.Listen(); err != nil {
		t.Fatal(err)
	}
	go server.Run()
	defer server.Close()

	setting.whoisServer = server.server.Addr().String()
	result := whois("AS6939")
	if !strings.Contains(result, "HURRICANE") {
		t.Errorf("Whois AS6939 failed, got %s", result)
	}
}

func TestWhoisWithoutServer(t *testing.T) {
	setting.whoisServer = ""
	result := whois("AS6939")
	if result != "" {
		t.Errorf("Whois AS6939 without server produced output, got %s", result)
	}
}

func TestWhoisConnectionError(t *testing.T) {
	setting.whoisServer = "127.0.0.1:1"
	result := whois("AS6939")
	if !strings.Contains(result, "connect: connection refused") {
		t.Errorf("Whois AS6939 without server produced output, got %s", result)
	}
}

func TestWhoisHostProcess(t *testing.T) {
	setting.whoisServer = "/bin/sh -c \"echo Mock Result\""
	result := whois("AS6939")
	if result != "Mock Result\n" {
		t.Errorf("Whois didn't produce expected result, got %s", result)
	}
}

func TestWhoisHostProcessMalformedCommand(t *testing.T) {
	setting.whoisServer = "/bin/sh -c \"mock"
	result := whois("AS6939")
	if result != "EOF found when expecting closing quote" {
		t.Errorf("Whois didn't produce expected result, got %s", result)
	}
}

func TestWhoisHostProcessError(t *testing.T) {
	setting.whoisServer = "/nonexistent"
	result := whois("AS6939")
	if !strings.Contains(result, "no such file or directory") {
		t.Errorf("Whois didn't produce expected result, got %s", result)
	}
}

func TestWhoisHostProcessVeryLong(t *testing.T) {
	setting.whoisServer = "/bin/sh -c \"for i in $(seq 1 131072); do printf 'A'; done\""
	result := whois("AS6939")
	if len(result) != 65535 {
		t.Errorf("Whois result incorrectly truncated, actual len %d", len(result))
	}
}

func TestWhoisIPv6(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response:      AS6939Response,
		listenAddr:    "[::1]:0",
	}

	if err := server.Listen(); err != nil {
		t.Skip("IPv6 not available:", err)
	}
	go server.Run()
	defer server.Close()

	setting.whoisServer = server.server.Addr().String()
	result := whois("AS6939")
	if !strings.Contains(result, "HURRICANE") {
		t.Errorf("Whois AS6939 over IPv6 failed, got %s", result)
	}
}

func TestWhoisIPv6WithoutPort(t *testing.T) {
	server := WhoisServer{
		t:             t,
		expectedQuery: "AS6939",
		response:      AS6939Response,
		listenAddr:    "[::1]:43",
	}

	if err := server.Listen(); err != nil {
		t.Skip("IPv6 not available or port 43 not bindable:", err)
	}
	go server.Run()
	defer server.Close()

	// Test that bare IPv6 address (without port) gets default port 43 appended
	setting.whoisServer = "::1"
	result := whois("AS6939")
	if !strings.Contains(result, "HURRICANE") {
		t.Errorf("Whois AS6939 over IPv6 (bare address) failed, got %s", result)
	}
}

func TestWhoisIPv6ConnectionError(t *testing.T) {
	// Use IPv6 loopback with a port that should be refused
	setting.whoisServer = "[::1]:1"
	result := whois("AS6939")
	if !strings.Contains(result, "connect: connection refused") && !strings.Contains(result, "network is unreachable") {
		t.Errorf("Whois AS6939 IPv6 connection error produced unexpected output, got %s", result)
	}
}
