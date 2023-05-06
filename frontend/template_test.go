package main

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestSummaryRowDataNameHasPrefix(t *testing.T) {
	data := SummaryRowData{
		Name: "mock",
	}

	assert.Equal(t, data.NameHasPrefix("m"), true)
	assert.Equal(t, data.NameHasPrefix("n"), false)
}

func TestSummaryRowDataNameContains(t *testing.T) {
	data := SummaryRowData{
		Name: "mock",
	}

	assert.Equal(t, data.NameContains("oc"), true)
	assert.Equal(t, data.NameContains("no"), false)
}
