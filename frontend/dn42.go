package main

import (
	"fmt"
	"strings"
)

func dn42WhoisFilter(whois string) string {
	commandResult := ""
	isNextSection := false
	skippedLines := 0

	// Filter out some long (and useless) keys
	filteredPrefix := []string{
		"descr:", "remarks:", "ds-rdata:", "auth:", "country:",
		"nserver:", "status:", "pgp-fingerprint:", "mp-import:", "mp-export:",
		"members:", "key:", "inetnum:", "inet6num:",
	}
	for _, s := range strings.Split(whois, "\n") {
		if len(s) == 0 {
			// This line is a separation between parts of results
			// Only keep last part of result
			isNextSection = true
			continue
		}
		shouldSkip := false
		for _, filtered := range filteredPrefix {
			if strings.HasPrefix(s, filtered) {
				shouldSkip = true
			}
		}
		if shouldSkip {
			skippedLines++
			continue
		}

		if isNextSection {
			isNextSection = false
			skippedLines = 0
			commandResult = ""
		}

		commandResult += s + "\n"
	}

	return commandResult + fmt.Sprintf("\n%d line(s) skipped.\n", skippedLines)
}
