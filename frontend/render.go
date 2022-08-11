package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// static options map
var optionsMap = map[string]string{
	"summary":                          "show protocols",
	"detail":                           "show protocols all ...",
	"route_from_protocol":              "show route protocol ...",
	"route_from_protocol_all":          "show route protocol ... all",
	"route_from_protocol_all_primary":  "show route protocol ... all primary",
	"route_filtered_from_protocol":     "show route filtered protocol ...",
	"route_filtered_from_protocol_all": "show route filtered protocol ... all",
	"route_from_origin":                "show route where bgp_path.last = ...",
	"route_from_origin_all":            "show route where bgp_path.last = ... all",
	"route_from_origin_all_primary":    "show route where bgp_path.last = ... all primary",
	"route":                            "show route for ...",
	"route_all":                        "show route for ... all",
	"route_bgpmap":                     "show route for ... (bgpmap)",
	"route_where":                      "show route where net ~ [ ... ]",
	"route_where_all":                  "show route where net ~ [ ... ] all",
	"route_where_bgpmap":               "show route where net ~ [ ... ] (bgpmap)",
	"route_generic":                    "show route ...",
	"generic":                          "show ...",
	"whois":                            "whois ...",
	"traceroute":                       "traceroute ...",
}

// pre-compiled regexp and constant statemap for summary rendering
var splitSummaryLine = regexp.MustCompile(`(\w+)(\s+)(\w+)(\s+)([\w-]+)(\s+)(\w+)(\s+)([0-9\-\. :]+)(.*)`)
var summaryStateMap = map[string]string{
	"up":      "success",
	"down":    "secondary",
	"start":   "danger",
	"passive": "info",
}

// render the page template
func renderPageTemplate(w http.ResponseWriter, r *http.Request, title string, content template.HTML) {
	path := r.URL.Path[1:]
	split := strings.SplitN(path, "/", 3)

	isWhois := strings.ToLower(split[0]) == "whois"
	whoisTarget := strings.Join(split[1:], "/")

	// Use a default URL if the request URL is too short
	// The URL is for return to summary page
	if len(split) < 2 {
		path = "summary/" + strings.Join(setting.servers, "+") + "/"
	} else if len(split) == 2 {
		path += "/"
	}

	split = strings.SplitN(path, "/", 3)

	args := TemplatePage{
		Options:              optionsMap,
		Servers:              setting.servers,
		ServersDisplay:       setting.serversDisplay,
		AllServersLinkActive: strings.EqualFold(split[1], strings.Join(setting.servers, "+")),
		AllServersURL:        strings.Join(setting.servers, "+"),
		AllServerTitle:       setting.navBarAllServer,
		AllServersURLCustom:  setting.navBarAllURL,
		IsWhois:              isWhois,
		WhoisTarget:          whoisTarget,

		URLOption:  strings.ToLower(split[0]),
		URLServer:  strings.ToLower(split[1]),
		URLCommand: split[2],
		Title:      setting.titleBrand + title,
		Brand:      setting.navBarBrand,
		BrandURL:   setting.navBarBrandURL,
		Content:    content,
	}

	tmpl := TemplateLibrary["page"]
	err := tmpl.Execute(w, args)
	if err != nil {
		fmt.Println("Error rendering page:", err.Error())
	}

}

// Write the given text to http response, and add whois links for
// ASNs and IP addresses
func smartFormatter(s string) template.HTML {
	var result string
	result += "<pre>"
	s = template.HTMLEscapeString(s)
	for _, line := range strings.Split(s, "\n") {
		var lineFormatted string
		if strings.HasPrefix(strings.TrimSpace(line), "BGP.as_path:") || strings.HasPrefix(strings.TrimSpace(line), "Neighbor AS:") || strings.HasPrefix(strings.TrimSpace(line), "Local AS:") {
			lineFormatted = regexp.MustCompile(`(\d+)`).ReplaceAllString(line, `<a href="/whois/AS${1}" class="whois">${1}</a>`)
		} else {
			lineFormatted = regexp.MustCompile(`([a-zA-Z0-9\-]*\.([a-zA-Z]{2,3}){1,2})(\s|$)`).ReplaceAllString(line, `<a href="/whois/${1}" class="whois">${1}</a>${3}`)
			lineFormatted = regexp.MustCompile(`\[AS(\d+)`).ReplaceAllString(lineFormatted, `[<a href="/whois/AS${1}" class="whois">AS${1}</a>`)
			lineFormatted = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`).ReplaceAllString(lineFormatted, `<a href="/whois/${1}" class="whois">${1}</a>`)
			lineFormatted = regexp.MustCompile(`(?i)(([a-f\d]{0,4}:){3,10}[a-f\d]{0,4})`).ReplaceAllString(lineFormatted, `<a href="/whois/${1}" class="whois">${1}</a>`)
		}
		result += lineFormatted + "\n"
	}
	result += "</pre>"
	return template.HTML(result)
}

// Parse bird show protocols result
func summaryParse(data string, serverName string) (TemplateSummary, error) {
	args := TemplateSummary{
		ServerName: serverName,
		Raw:        data,
	}

	lines := strings.Split(strings.TrimSpace(data), "\n")
	if len(lines) <= 1 {
		// Likely backend returned an error message
		return args, errors.New(strings.TrimSpace(data))
	}

	// extract the table header
	for _, col := range strings.Split(lines[0], " ") {
		colTrimmed := strings.TrimSpace(col)
		if len(colTrimmed) == 0 {
			continue
		}
		args.Header = append(args.Header, col)
	}

	// Build regexp for nameFilter
	nameFilterRegexp := regexp.MustCompile(setting.nameFilter)

	// sort the remaining rows
	rows := lines[1:]
	sort.Strings(rows)

	// parse each line
	for _, line := range rows {

		// Ignore empty lines
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parse a total of 6 columns from bird summary
		lineSplitted := splitSummaryLine.FindStringSubmatch(line)
		if lineSplitted == nil {
			continue
		}

		var row SummaryRowData

		if len(lineSplitted) >= 2 {
			row.Name = strings.TrimSpace(lineSplitted[1])
			if setting.nameFilter != "" && nameFilterRegexp.MatchString(row.Name) {
				continue
			}
		}
		if len(lineSplitted) >= 4 {
			row.Proto = strings.TrimSpace(lineSplitted[3])
			// Filter away unwanted protocol types, if setting.protocolFilter is non-empty
			found := false
			for _, protocol := range setting.protocolFilter {
				if strings.EqualFold(row.Proto, protocol) {
					found = true
					break
				}
			}

			if len(setting.protocolFilter) > 0 && !found {
				continue
			}
		}
		if len(lineSplitted) >= 6 {
			row.Table = strings.TrimSpace(lineSplitted[5])
		}
		if len(lineSplitted) >= 8 {
			row.State = strings.TrimSpace(lineSplitted[7])
			row.MappedState = summaryStateMap[row.State]
		}
		if len(lineSplitted) >= 10 {
			row.Since = strings.TrimSpace(lineSplitted[9])
		}
		if len(lineSplitted) >= 11 {
			row.Info = strings.TrimSpace(lineSplitted[10])
		}

		// add to the result
		args.Rows = append(args.Rows, row)
	}

	return args, nil
}

// Output a table for the summary page
func summaryTable(data string, serverName string) template.HTML {
	result, err := summaryParse(data, serverName)

	if err != nil {
		return template.HTML("<pre>" + template.HTMLEscapeString(err.Error()) + "</pre>")
	}

	// render the summary template
	tmpl := TemplateLibrary["summary"]
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, result)
	if err != nil {
		fmt.Println("Error rendering summary:", err.Error())
	}

	return template.HTML(buffer.String())
}
