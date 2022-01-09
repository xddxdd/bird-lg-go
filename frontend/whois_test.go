package main

import (
	"strings"
	"testing"
)

func TestWhois(t *testing.T) {
	checkNetwork(t)

	setting.whoisServer = "whois.arin.net"
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
