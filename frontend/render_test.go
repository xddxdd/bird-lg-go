package main

import (
	"html/template"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

const BirdSummaryData = `Name       Proto      Table      State  Since         Info
static1    Static     master4    up     2021-08-27
static2    Static     master6    up     2021-08-27
device1    Device     ---        up     2021-08-27
kernel1    Kernel     master6    up     2021-08-27
kernel2    Kernel     master4    up     2021-08-27
direct1    Direct     ---        up     2021-08-27
int_babel  Babel      ---        up     2021-08-27
`

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
	renderPageTemplate(w, r, title, template.HTML(content))

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

	r := httptest.NewRequest("GET", "/whois/"+evil, nil)
	w := httptest.NewRecorder()

	// renderPageTemplate doesn't escape content, filter is done beforehand
	renderPageTemplate(w, r, evil, "Test Content")

	resultBytes, _ := ioutil.ReadAll(w.Result().Body)
	result := string(resultBytes)

	if strings.Contains(result, evil) {
		t.Errorf("XSS injection succeeded: %s", result)
	}
}

// https://github.com/xddxdd/bird-lg-go/issues/57
func TestRenderPageTemplateXSS_2(t *testing.T) {
	initSettings()

	evil := "<script>alert('evil');</script>"

	r := httptest.NewRequest("GET", "/generic/dummy_server/"+evil, nil)
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
	result := string(smartFormatter(evil))

	if strings.Contains(result, evil) {
		t.Errorf("XSS injection succeeded: %s", result)
	}
}

func TestSummaryTableXSS(t *testing.T) {
	evil := "<script>alert('evil');</script>"
	evilData := `Name       Proto      Table      State  Since         Info
` + evil + ` ` + evil + `       ---        up     2021-01-04 17:21:44  ` + evil

	result := string(summaryTable(evilData, evil))

	if strings.Contains(result, evil) {
		t.Errorf("XSS injection succeeded: %s", result)
	}
}

func TestSummaryTableProtocolFilter(t *testing.T) {
	initSettings()
	setting.protocolFilter = []string{"Static", "Direct", "Babel"}

	result := string(summaryTable(BirdSummaryData, "testserver"))
	expectedInclude := []string{"static1", "static2", "int_babel", "direct1"}
	expectedExclude := []string{"device1", "kernel1", "kernel2"}

	for _, item := range expectedInclude {
		if !strings.Contains(result, item) {
			t.Errorf("Did not find expected %s in summary table output", result)
		}
	}
	for _, item := range expectedExclude {
		if strings.Contains(result, item) {
			t.Errorf("Found unexpected %s in summary table output", result)
		}
	}

	t.Cleanup(func() {
		setting.protocolFilter = []string{}
	})
}

func TestSummaryTableNameFilter(t *testing.T) {
	initSettings()
	setting.nameFilter = "^static"

	result := string(summaryTable(BirdSummaryData, "testserver"))
	expectedInclude := []string{"device1", "kernel1", "kernel2", "direct1", "int_babel"}
	expectedExclude := []string{"static1", "static2"}

	for _, item := range expectedInclude {
		if !strings.Contains(result, item) {
			t.Errorf("Did not find expected %s in summary table output", result)
		}
	}
	for _, item := range expectedExclude {
		if strings.Contains(result, item) {
			t.Errorf("Found unexpected %s in summary table output", result)
		}
	}

	t.Cleanup(func() {
		setting.nameFilter = ""
	})
}
