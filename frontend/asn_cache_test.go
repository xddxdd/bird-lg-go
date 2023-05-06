package main

import (
	"strings"
	"testing"
)

func TestGetASNRepresentationDNS(t *testing.T) {
	checkNetwork(t)

	setting.dnsInterface = "asn.cymru.com"
	setting.whoisServer = ""
	cache := make(ASNCache)
	result := cache.Lookup("6939")
	if !strings.Contains(result, "HURRICANE") {
		t.Errorf("Lookup AS6939 failed, got %s", result)
	}
}

func TestGetASNRepresentationWhois(t *testing.T) {
	checkNetwork(t)

	setting.dnsInterface = ""
	setting.whoisServer = "whois.arin.net"
	cache := make(ASNCache)
	result := cache.Lookup("6939")
	if !strings.Contains(result, "HURRICANE") {
		t.Errorf("Lookup AS6939 failed, got %s", result)
	}
}

func TestGetASNRepresentationFallback(t *testing.T) {
	setting.dnsInterface = ""
	setting.whoisServer = ""
	cache := make(ASNCache)
	result := cache.Lookup("6939")
	if result != "AS6939" {
		t.Errorf("Lookup AS6939 failed, got %s", result)
	}
}
