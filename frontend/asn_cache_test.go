package main

import (
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
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

func TestGetASNRepresentationDNSFallback(t *testing.T) {
	checkNetwork(t)

	setting.dnsInterface = "invalid.example.com"
	setting.whoisServer = "whois.arin.net"
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
	assert.Equal(t, result, "AS6939")
}
