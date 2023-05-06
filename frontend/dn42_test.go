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

func TestShortenWhoisFilterShorterMode(t *testing.T) {
	input := `
Information line that will be removed

# Comment that will be removed
Name: Redacted for privacy
Descr: This is a vvvvvvvvvvvvvvvvvvvvvvveeeeeeeeeeeeeeeeeeeerrrrrrrrrrrrrrrrrrrrrrrryyyyyyyyyyyyyyyyyyy long line that will be skipped.
Looooooooooooooooooooooong key: this line will be skipped.

Preserved1: this line isn't removed.
Preserved2: this line isn't removed.
Preserved3: this line isn't removed.
Preserved4: this line isn't removed.
Preserved5: this line isn't removed.

`

	result := shortenWhoisFilter(input)

	expectedResult := `Preserved1: this line isn't removed.
Preserved2: this line isn't removed.
Preserved3: this line isn't removed.
Preserved4: this line isn't removed.
Preserved5: this line isn't removed.

3 line(s) skipped.
`

	if result != expectedResult {
		t.Errorf("Output doesn't match expected: %s", result)
	}
}

func TestShortenWhoisFilterLongerMode(t *testing.T) {
	input := `
Information line that will be removed

# Comment that will be removed
Name: Redacted for privacy
Descr: This is a vvvvvvvvvvvvvvvvvvvvvvveeeeeeeeeeeeeeeeeeeerrrrrrrrrrrrrrrrrrrrrrrryyyyyyyyyyyyyyyyyyy long line that will be skipped.
Looooooooooooooooooooooong key: this line will be skipped.

Preserved1: this line isn't removed.

`

	result := shortenWhoisFilter(input)

	expectedResult := `Information line that will be removed
Descr: This is a vvvvvvvvvvvvvvvvvvvvvvveeeeeeeeeeeeeeeeeeeerrrrrrrrrrrrrrrrrrrrrrrryyyyyyyyyyyyyyyyyyy long line that will be skipped.
Looooooooooooooooooooooong key: this line will be skipped.
Preserved1: this line isn't removed.

7 line(s) skipped.
`

	if result != expectedResult {
		t.Errorf("Output doesn't match expected: %s", result)
	}
}

func TestShortenWhoisFilterSkipNothing(t *testing.T) {
	input := `Preserved1: this line isn't removed.
Preserved2: this line isn't removed.
Preserved3: this line isn't removed.
Preserved4: this line isn't removed.
Preserved5: this line isn't removed.
`

	result := shortenWhoisFilter(input)

	if result != input {
		t.Errorf("Output doesn't match expected: %s", result)
	}
}
