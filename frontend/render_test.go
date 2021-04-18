package main

import (
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func initSettings() {
	setting.servers = []string{"alpha"}
	setting.serversDisplay = []string{"alpha"}
	setting.titleBrand = "Bird-lg Go"
	setting.navBarBrand = "Bird-lg Go"

	ImportTemplates()
}

func TestRenderPageTemplate(t *testing.T) {
	initSettings()

	title := "Test Title"
	content := "Test Content"

	r := httptest.NewRequest("GET", "/route/alpha/192.168.0.1/", nil)
	w := httptest.NewRecorder()
	renderPageTemplate(w, r, title, content)

	resultBytes, _ := ioutil.ReadAll(w.Result().Body)
	result := string(resultBytes)

	if !strings.Contains(result, title) {
		t.Error("Title not found in output")
	}
	if !strings.Contains(result, content) {
		t.Error("Content not found in output")
	}
}

func TestRenderPageTemplateXSS(t *testing.T) {
	initSettings()

	evil := "<script>alert('evil');</script>"

	r := httptest.NewRequest("GET", "/whois/"+url.PathEscape(evil), nil)
	w := httptest.NewRecorder()

	// renderPageTemplate doesn't escape content, filter is done beforehand
	renderPageTemplate(w, r, evil, "Test Content")

	resultBytes, _ := ioutil.ReadAll(w.Result().Body)
	result := string(resultBytes)

	if strings.Contains(result, evil) {
		t.Errorf("XSS injection succeeded: %s", result)
	}
}

func TestSmartFormatterXSS(t *testing.T) {
	evil := "<script>alert('evil');</script>"
	result := smartFormatter(evil)

	if strings.Contains(result, evil) {
		t.Errorf("XSS injection succeeded: %s", result)
	}
}

func TestSummaryTableXSS(t *testing.T) {
	evil := "<script>alert('evil');</script>"
	evilData := `Name       Proto      Table      State  Since         Info
` + evil + ` ` + evil + `       ---        up     2021-01-04 17:21:44  ` + evil

	result := summaryTable(evilData, evil)

	if strings.Contains(result, evil) {
		t.Errorf("XSS injection succeeded: %s", result)
	}
}
