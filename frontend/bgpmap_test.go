package main

import (
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"testing"
)

func readDataFile(t *testing.T, filename string) string {
	_, sourceName, _, _ := runtime.Caller(0)
	projectRoot := path.Join(path.Dir(sourceName), "..")
	dir := path.Join(projectRoot, filename)

	data, err := ioutil.ReadFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
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

func TestBirdRouteToGraph(t *testing.T) {
	setting.dnsInterface = ""

	input := readDataFile(t, "frontend/test_data/bgpmap_case1.txt")
	result := birdRouteToGraph([]string{"node"}, []string{input}, "target")

	// Source node must exist
	if result.GetPoint("node") == nil {
		t.Error("Result doesn't contain point node")
	}
	// Last hop must exist
	if result.GetPoint("4242423914") == nil {
		t.Error("Result doesn't contain point 4242423914")
	}
	// Destination must exist
	if result.GetPoint("target") == nil {
		t.Error("Result doesn't contain point target")
	}

	// Verify that a few paths exist
	if result.GetEdge("node", "4242423914") == nil {
		t.Error("Result doesn't contain edge from node to 4242423914")
	}
	if result.GetEdge("node", "4242422688") == nil {
		t.Error("Result doesn't contain edge from node to 4242422688")
	}
	if result.GetEdge("4242422688", "4242423914") == nil {
		t.Error("Result doesn't contain edge from 4242422688 to 4242423914")
	}
	if result.GetEdge("4242423914", "target") == nil {
		t.Error("Result doesn't contain edge from 4242423914 to target")
	}
}

func TestBirdRouteToGraphviz(t *testing.T) {
	setting.dnsInterface = ""

	input := readDataFile(t, "frontend/test_data/bgpmap_case1.txt")
	result := birdRouteToGraphviz([]string{"node"}, []string{input}, "target")

	if !strings.Contains(result, "digraph {") {
		t.Error("Response is not Graphviz data")
	}
}
