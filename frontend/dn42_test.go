package main

import (
	"testing"
)

func TestDN42WhoisFilter(t *testing.T) {
	input := "name: Testing\ndescr: Description"

	result := dn42WhoisFilter(input)

	expectedResult := `name: Testing

1 line(s) skipped.
`

	if result != expectedResult {
		t.Errorf("Output doesn't match expected: %s", result)
	}
}

func TestDN42WhoisFilterUnneeded(t *testing.T) {
	input := "name: Testing\nwhatever: Description"

	result := dn42WhoisFilter(input)

	if result != input+"\n" {
		t.Errorf("Output doesn't match expected: %s", result)
	}
}
