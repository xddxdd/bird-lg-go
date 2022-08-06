package main

import (
	"embed"
	"html/template"
	"strings"
)

// import templates and other assets
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
		template, err := template.New(tmpl).Parse(string(def))
		if err != nil {
			panic("Unable to parse template (" + TEMPLATE_PATH + tmpl + ": " + err.Error())
		}

		// store in the library
		TemplateLibrary[tmpl] = template
	}

}
