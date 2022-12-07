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
// func TestBirdRouteToGraphviz(t *testing.T) {
// 	setting.dnsInterface = ""

// 	// Don't change formatting of the following strings!

// 	fakeResult := `192.168.0.1/32       unicast [alpha 2021-01-14 from 192.168.0.2] * (100) [AS12345i]
// 	via 192.168.0.2 on eth0
// 	Type: BGP univ
// 	BGP.origin: IGP
// 	BGP.as_path: 4242422601
// 	BGP.next_hop: 172.18.0.2`

// 	expectedResult := strings.Split(`digraph {
// "AS4242422601" ["color"="red"];
// "AS4242422601" -> "Target: 192.168.0.1" ["color"="red"];
// "Target: 192.168.0.1" ["shape"="diamond","color"="red"];
// "alpha" ["color"="blue","shape"="box"];
// "alpha" -> "AS4242422601" ["fontsize"="12.0","color"="red","label"="alpha*\n172.18.0.2"];
// }
// `, "\n")

// 	result := birdRouteToGraphviz([]string{
// 		"alpha",
// 	}, []string{
// 		fakeResult,
// 	}, "192.168.0.1")


// 	for _, line := range strings.Split(result, "\n") {
// 		println(line)
// 		if !contains(expectedResult, line) {
// 			t.Errorf("Unexpected line in result: %s", line)
// 		} else {
// 			println("OK")
// 		}
// 	}
// }

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
