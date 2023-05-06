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
		"members:", "key:", "inetnum:", "inet6num:", " ", "+",
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

	if skippedLines > 0 {
		return commandResult + fmt.Sprintf("\n%d line(s) skipped.\n", skippedLines)
	} else {
		return commandResult
	}
}

/* experimental, behavior may change */
func shortenWhoisFilter(whois string) string {
	commandResult := ""
	commandResultLonger := ""
	lines := 0
	linesLonger := 0
	skippedLines := 0
	skippedLinesLonger := 0

	for _, s := range strings.Split(whois, "\n") {
		s = strings.TrimSpace(s)

		shouldSkip := false
		shouldSkip = shouldSkip || len(s) == 0
		shouldSkip = shouldSkip || len(s) > 0 && s[0] == '#'
		shouldSkip = shouldSkip || strings.Contains(strings.ToUpper(s), "REDACTED")

		if shouldSkip {
			skippedLinesLonger++
			continue
		}

		commandResultLonger += s + "\n"
		linesLonger++

		shouldSkip = shouldSkip || len(s) > 80
		shouldSkip = shouldSkip || !strings.Contains(s, ":")
		shouldSkip = shouldSkip || strings.Index(s, ":") > 20

		if shouldSkip {
			skippedLines++
			continue
		}

		commandResult += s + "\n"
		lines++
	}

	if lines < 5 {
		commandResult = commandResultLonger
		skippedLines = skippedLinesLonger
	}

	if skippedLines > 0 {
		return commandResult + fmt.Sprintf("\n%d line(s) skipped.\n", skippedLines)
	} else {
		return commandResult
	}
}
