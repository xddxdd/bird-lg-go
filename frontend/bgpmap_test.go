package main

import (
	"strings"
	"testing"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func TestGetASNRepresentationDNS(t *testing.T) {
	checkNetwork(t)

	setting.dnsInterface = "asn.cymru.com"
	setting.whoisServer = ""
	result := getASNRepresentation("6939")
	if !strings.Contains(result, "HURRICANE") {
		t.Errorf("Lookup AS6939 failed, got %s", result)
	}
}

func TestGetASNRepresentationWhois(t *testing.T) {
	checkNetwork(t)

	setting.dnsInterface = ""
	setting.whoisServer = "whois.arin.net"
	result := getASNRepresentation("6939")
	if !strings.Contains(result, "HURRICANE") {
		t.Errorf("Lookup AS6939 failed, got %s", result)
	}
}

func TestGetASNRepresentationFallback(t *testing.T) {
	setting.dnsInterface = ""
	setting.whoisServer = ""
	result := getASNRepresentation("6939")
	if result != "AS6939" {
		t.Errorf("Lookup AS6939 failed, got %s", result)
	}
}

// Broken due to random order of attributes
func TestBirdRouteToGraphviz(t *testing.T) {
	setting.dnsInterface = ""

	// Don't change formatting of the following strings!

	fakeResult := `192.168.0.1/32       unicast [alpha 2021-01-14 from 192.168.0.2] * (100) [AS12345i]
	via 192.168.0.2 on eth0
	Type: BGP univ
	BGP.origin: IGP
	BGP.as_path: 4242422601
	BGP.next_hop: 172.18.0.2`

	expectedLinesInResult := []string{
		`"AS4242422601" [`,
		`"AS4242422601" -> "Target: 192.168.0.1" [`,
		`"Target: 192.168.0.1" [`,
		`"alpha" [`,
		`"alpha" -> "AS4242422601" [`,
	}

	result := birdRouteToGraphviz([]string{
		"alpha",
	}, []string{
		fakeResult,
	}, "192.168.0.1")


	for _, line := range expectedLinesInResult {
		if !strings.Contains(result, line) {
			t.Errorf("Expected line in result not found: %s", line)
		}
	}
}

func TestBirdRouteToGraphvizXSS(t *testing.T) {
	setting.dnsInterface = ""

	// Don't change formatting of the following strings!

	fakeResult := `<script>alert("evil!")</script>`

	result := birdRouteToGraphviz([]string{
		"alpha",
	}, []string{
		fakeResult,
	}, fakeResult)

	if strings.Contains(result, "<script>") {
		t.Errorf("XSS injection succeeded: %s", result)
	}
}
