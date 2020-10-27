package main

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func renderTemplate(w http.ResponseWriter, r *http.Request, title string, content string) {
	path := r.URL.Path[1:]
	split := strings.SplitN(path, "/", 4)

	isWhois := strings.ToLower(split[0]) == "whois"
	whoisTarget := strings.Join(split[1:], "/")

	// Use a default URL if the request URL is too short
	// The URL is for return to IPv4 summary page
	if len(split) < 3 {
		path = "ipv4/summary/" + strings.Join(setting.servers, "+") + "/"
	} else if len(split) == 3 {
		path += "/"
	}

	split = strings.SplitN(path, "/", 4)

	var args tmplArguments
	args.Options = map[string]string{
		"summary":            "show protocols",
		"detail":             "show protocols all",
		"route":              "show route for ...",
		"route_all":          "show route for ... all",
		"route_bgpmap":       "show route for ... (bgpmap)",
		"route_where":        "show route where net ~ [ ... ]",
		"route_where_all":    "show route where net ~ [ ... ] all",
		"route_where_bgpmap": "show route where net ~ [ ... ] (bgpmap)",
		"route_generic":      "show route ...",
		"generic":            "show ...",
		"whois":              "whois ...",
		"traceroute":         "traceroute ...",
	}
	args.Servers = setting.servers
	args.AllServersLinkActive = strings.ToLower(split[2]) == strings.ToLower(strings.Join(setting.servers, "+"))
	args.AllServersURL = strings.Join(setting.servers, "+")
	args.IsWhois = isWhois
	args.WhoisTarget = whoisTarget

	args.URLProto = strings.ToLower(split[0])
	args.URLOption = strings.ToLower(split[1])
	args.URLServer = strings.ToLower(split[2])
	args.URLCommand = split[3]

	args.Title = setting.titleBrand + title
	args.Brand = setting.navBarBrand
	args.Content = content

	err := tmpl.Execute(w, args)
	if err != nil {
		panic(err)
	}
}

// Write the given text to http response, and add whois links for
// ASNs and IP addresses
func smartFormatter(s string) string {
	var result string
	result += "<pre>"
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
	return result
}

type summaryTableArguments struct {
	Headers []string
	Lines   [][]string
}

// Output a table for the summary page
func summaryTable(isIPv6 bool, data string, serverName string) string {
	var result string

	// Sort the table, excluding title row
	stringsSplitted := strings.Split(strings.TrimSpace(data), "\n")
	if len(stringsSplitted) <= 1 {
		// Likely backend returned an error message
		result = "<pre>" + strings.TrimSpace(data) + "</pre>"
	} else {
		// Draw the table head
		result += `<table class="table table-striped table-bordered table-sm">`
		result += `<thead>`
		for _, col := range strings.Split(stringsSplitted[0], " ") {
			colTrimmed := strings.TrimSpace(col)
			if len(colTrimmed) == 0 {
				continue
			}
			result += `<th scope="col">` + colTrimmed + `</th>`
		}
		result += `</thead><tbody>`

		stringsWithoutTitle := stringsSplitted[1:]
		sort.Strings(stringsWithoutTitle)

		for _, line := range stringsWithoutTitle {
			// Ignore empty lines
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			// Parse a total of 6 columns from bird summary
			lineSplitted := regexp.MustCompile(`(\w+)(\s+)(\w+)(\s+)([\w-]+)(\s+)(\w+)(\s+)([0-9\-\. :]+)(.*)`).FindStringSubmatch(line)
			if lineSplitted == nil {
				continue
			}

			var row [6]string
			if len(lineSplitted) >= 2 {
				row[0] = strings.TrimSpace(lineSplitted[1])
			}
			if len(lineSplitted) >= 4 {
				row[1] = strings.TrimSpace(lineSplitted[3])
			}
			if len(lineSplitted) >= 6 {
				row[2] = strings.TrimSpace(lineSplitted[5])
			}
			if len(lineSplitted) >= 8 {
				row[3] = strings.TrimSpace(lineSplitted[7])
			}
			if len(lineSplitted) >= 10 {
				row[4] = strings.TrimSpace(lineSplitted[9])
			}
			if len(lineSplitted) >= 11 {
				row[5] = strings.TrimSpace(lineSplitted[10])
			}

			// Draw the row in red if the link isn't up
			result += `<tr class="` + (map[string]string{
				"up":      "table-success",
				"down":    "table-secondary",
				"start":   "table-danger",
				"passive": "table-info",
			})[row[3]] + `">`
			// Add link to detail for first column
			if isIPv6 {
				result += `<td><a href="/ipv6/detail/` + serverName + `/` + row[0] + `">` + row[0] + `</a></td>`
			} else {
				result += `<td><a href="/ipv4/detail/` + serverName + `/` + row[0] + `">` + row[0] + `</a></td>`
			}
			// Draw the other cells
			for i := 1; i < 6; i++ {
				result += "<td>" + row[i] + "</td>"
			}
			result += "</tr>"
		}
		result += "</tbody></table>"
		result += "<!--" + data + "-->"
	}

	return result
}
