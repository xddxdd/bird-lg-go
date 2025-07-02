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

func TestSummaryRowDataFromLine(t *testing.T) {
	data := SummaryRowDataFromLine("sys_device Device     ---        up     2025-06-27 21:23:08")

	assert.Equal(t, data.Name, "sys_device")
	assert.Equal(t, data.Proto, "Device")
	assert.Equal(t, data.Table, "---")
	assert.Equal(t, data.State, "up")
	assert.Equal(t, data.Since, "2025-06-27 21:23:08")
}

func TestSummaryRowDataFromLineNumeric(t *testing.T) {
	data := SummaryRowDataFromLine("12345 Device     ---        up     2025-06-27 21:23:08")

	assert.Equal(t, data.Name, "12345")
	assert.Equal(t, data.Proto, "Device")
	assert.Equal(t, data.Table, "---")
	assert.Equal(t, data.State, "up")
	assert.Equal(t, data.Since, "2025-06-27 21:23:08")
}

func TestSummaryRowDataFromLinePipe(t *testing.T) {
	data := SummaryRowDataFromLine("pipe Pipe       ---        up     2025-06-27 21:23:08  master4 <=> pipe_v4")

	assert.Equal(t, data.Name, "pipe")
	assert.Equal(t, data.Proto, "Pipe")
	assert.Equal(t, data.Table, "---")
	assert.Equal(t, data.State, "up")
	assert.Equal(t, data.Since, "2025-06-27 21:23:08")
	assert.Equal(t, data.Info, "master4 <=> pipe_v4")
}

func TestSummaryRowDataFromLineBGP(t *testing.T) {
	data := SummaryRowDataFromLine("bgp BGP        ---        up     2025-06-30 20:45:33  Established")

	assert.Equal(t, data.Name, "bgp")
	assert.Equal(t, data.Proto, "BGP")
	assert.Equal(t, data.Table, "---")
	assert.Equal(t, data.State, "up")
	assert.Equal(t, data.Since, "2025-06-30 20:45:33")
	assert.Equal(t, data.Info, "Established")
}

func TestSummaryRowDataFromLineBGPPassive(t *testing.T) {
	data := SummaryRowDataFromLine("passive   BGP        ---        start  2025-06-27 21:23:08  Passive")

	assert.Equal(t, data.Name, "passive")
	assert.Equal(t, data.Proto, "BGP")
	assert.Equal(t, data.Table, "---")
	assert.Equal(t, data.State, "start")
	assert.Equal(t, data.Since, "2025-06-27 21:23:08")
	assert.Equal(t, data.Info, "Passive")
}

func TestSummaryRowDataFromLineWithDash(t *testing.T) {
	data := SummaryRowDataFromLine("ibgp_test-01 BGP        ---        up     07:16:51.656  Established")

	assert.Equal(t, data.Name, "ibgp_test-01")
	assert.Equal(t, data.Proto, "BGP")
	assert.Equal(t, data.Table, "---")
	assert.Equal(t, data.State, "up")
	assert.Equal(t, data.Since, "07:16:51.656")
	assert.Equal(t, data.Info, "Established")
}
