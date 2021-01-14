package main

import (
	"strings"
	"text/template"
)

// template argument structures

// page
type TemplatePage struct {
	// Global options
	Options map[string]string
	Servers []string

	// Parameters related to current request
	AllServersLinkActive bool
	AllServersURL        string

	// Whois specific handling (for its unique URL)
	IsWhois     bool
	WhoisTarget string

	URLOption  string
	URLServer  string
	URLCommand string

	// Generated content to be displayed
	Title   string
	Brand   string
	Content string
}

// summary

type SummaryRowData struct {
	Name        string
	Proto       string
	Table       string
	State       string
	MappedState string
	Since       string
	Info        string
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
	Result string
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
	Result     string
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

// import templates from bindata

func ImportTemplates() {

	// create a new (blank) initial template
	TemplateLibrary = make(map[string]*template.Template)

	// for each template that is needed
	for _, tmpl := range requiredTemplates {

		// extract the template definition from the bindata
		def := MustAssetString("templates/" + tmpl + ".tpl")

		// and add it to the template library
		template, err := template.New(tmpl).Parse(def)
		if err != nil {
			panic("Unable to parse template (templates/" + tmpl + ": " + err.Error())
		}

		// store in the library
		TemplateLibrary[tmpl] = template
	}

}
