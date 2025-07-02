package main

import (
	"embed"
	"html/template"
	"net/url"
	"regexp"
	"strings"
)

// import templates and other assets
//
//go:embed assets
var assets embed.FS

const TEMPLATE_PATH = "assets/templates/"

// template argument structures

// page
type TemplatePage struct {
	// Global options
	Options        map[string]string
	Servers        []string
	ServersDisplay []string

	// Parameters related to current request
	AllServersLinkActive bool
	AllServerTitle       string
	AllServersURL        string
	AllServersURLCustom  string

	// Whois specific handling (for its unique URL)
	IsWhois     bool
	WhoisTarget string

	URLOption  string
	URLServer  string
	URLCommand string

	// Generated content to be displayed
	Title    string
	Brand    string
	BrandURL string
	Content  template.HTML
}

// summary
type SummaryRowData struct {
	Name        string `json:"name"`
	Proto       string `json:"proto"`
	Table       string `json:"table"`
	State       string `json:"state"`
	MappedState string `json:"-"`
	Since       string `json:"since"`
	Info        string `json:"info"`
}

// utility functions to allow filtering of results in the template

func (r SummaryRowData) NameHasPrefix(prefix string) bool {
	return strings.HasPrefix(r.Name, prefix)
}

func (r SummaryRowData) NameContains(prefix string) bool {
	return strings.Contains(r.Name, prefix)
}

func (r SummaryRowData) ProtocolMatches(protocols []string) bool {
	for _, protocol := range protocols {
		if strings.EqualFold(r.Proto, protocol) {
			return true
		}
	}
	return false
}

// pre-compiled regexp and constant statemap for summary rendering
var splitSummaryLine = regexp.MustCompile(`^([\w-]+)\s+(\w+)\s+([\w-]+)\s+(\w+)\s+([0-9\-\. :]+)(.*)$`)
var summaryStateMap = map[string]string{
	"up":      "success",
	"down":    "secondary",
	"start":   "danger",
	"passive": "info",
}

func SummaryRowDataFromLine(line string) *SummaryRowData {
	lineSplitted := splitSummaryLine.FindStringSubmatch(line)
	if lineSplitted == nil {
		return nil
	}

	var row SummaryRowData
	row.Name = strings.TrimSpace(lineSplitted[1])
	row.Proto = strings.TrimSpace(lineSplitted[2])
	row.Table = strings.TrimSpace(lineSplitted[3])
	row.State = strings.TrimSpace(lineSplitted[4])
	row.Since = strings.TrimSpace(lineSplitted[5])
	row.Info = strings.TrimSpace(lineSplitted[6])

	if strings.Contains(row.Info, "Passive") {
		row.MappedState = summaryStateMap["passive"]
	} else {
		row.MappedState = summaryStateMap[row.State]
	}

	return &row
}

type TemplateSummary struct {
	ServerName string
	Raw        string
	Header     []string
	Rows       []SummaryRowData
}

// whois
type TemplateWhois struct {
	Target string
	Result template.HTML
}

// bgpmap
type TemplateBGPmap struct {
	Servers []string
	Target  string
	Result  string
}

// bird
type TemplateBird struct {
	ServerName string
	Target     string
	Result     template.HTML
}

// global variable to hold the templates

var TemplateLibrary map[string]*template.Template

// list of required templates

var requiredTemplates = [...]string{
	"page",
	"summary",
	"whois",
	"bgpmap",
	"bird",
}

// define functions to be made available in templates

var funcMap = template.FuncMap{
	"pathescape": url.PathEscape,
}

// import templates from embedded assets

func ImportTemplates() {

	// create a new (blank) initial template
	TemplateLibrary = make(map[string]*template.Template)

	// for each template that is needed
	for _, tmpl := range requiredTemplates {

		// extract the template definition from the embedded assets
		def, err := assets.ReadFile(TEMPLATE_PATH + tmpl + ".tpl")
		if err != nil {
			panic("Unable to read template (" + TEMPLATE_PATH + tmpl + ": " + err.Error())
		}

		// and add it to the template library
		template, err := template.New(tmpl).Funcs(funcMap).Parse(string(def))
		if err != nil {
			panic("Unable to parse template (" + TEMPLATE_PATH + tmpl + ": " + err.Error())
		}

		// store in the library
		TemplateLibrary[tmpl] = template
	}

}
