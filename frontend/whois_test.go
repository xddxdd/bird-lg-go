package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
)

type WhoisServer struct {
	t             *testing.T
	expectedQuery string
	response      string
	server        net.Listener
}

const AS6939Response = `
ASNumber:       6939
ASName:         HURRICANE
ASHandle:       AS6939
RegDate:        1996-06-28
Updated:        2003-11-04
Ref:            https://rdap.arin.net/registry/autnum/6939
`

func (s *WhoisServer) Listen() {
	var err error
	s.server, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		s.t.Error(err)
	}
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

	server.Listen()
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
	setting.whoisServer = "127.0.0.1:0"
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
