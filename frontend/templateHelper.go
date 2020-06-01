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
		"summary":            "show protocol",
		"detail":             "show protocol all",
		"route":              "show route for ...",
		"route_all":          "show route for ... all",
		"route_bgpmap":       "show route for ... (bgpmap)",
		"route_where":        "show route where net ~ [ ... ]",
		"route_where_all":    "show route where net ~ [ ... ] all",
		"route_where_bgpmap": "show route where net ~ [ ... ] (bgpmap)",
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

	args.Title = title
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
			lineFormatted = regexp.MustCompile(`(\d+)`).ReplaceAllString(line, `<a href="/whois/${1}" class="whois">${1}</a>`)
		} else {
			lineFormatted = regexp.MustCompile(`([a-zA-Z0-9\-]*\.([a-zA-Z]{2,3}){1,2})(\s|$)`).ReplaceAllString(line, `<a href="/whois/${1}" class="whois">${1}</a>${3}`)
			lineFormatted = regexp.MustCompile(`\[AS(\d+)`).ReplaceAllString(lineFormatted, `[<a href="/whois/${1}" class="whois">AS${1}</a>`)
			lineFormatted = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`).ReplaceAllString(lineFormatted, `<a href="/whois/${1}" class="whois">${1}</a>`)
			lineFormatted = regexp.MustCompile(`(?i)(([a-f\d]{0,4}:){3,10}[a-f\d]{0,4})`).ReplaceAllString(lineFormatted, `<a href="/whois/${1}" class="whois">${1}</a>`)
		}
		result += lineFormatted + "\n"
	}
	result += "</pre>"
	return result
}

// Output a table for the summary page
func summaryTable(isIPv6 bool, data string, serverName string) string {
	var result string

	// Sort the table, excluding title row
	stringsSplitted := strings.Split(data, "\n")
	if len(stringsSplitted) > 1 {
		stringsWithoutTitle := stringsSplitted[1:]
		sort.Strings(stringsWithoutTitle)
		data = stringsSplitted[0] + "\n" + strings.Join(stringsWithoutTitle, "\n")
	}

	result += "<table class=\"table table-striped table-bordered table-sm\">"
	for lineID, line := range strings.Split(data, "\n") {
		var row [6]string
		var rowIndex int = 0

		words := strings.Split(line, " ")
		for wordID, word := range words {
			if len(word) == 0 {
				continue
			}
			if rowIndex < 4 {
				row[rowIndex] += word
				rowIndex++
			} else if len(words[wordID-1]) == 0 && rowIndex < len(row)-1 {
				if len(row[rowIndex]) > 0 {
					rowIndex++
				}
				row[rowIndex] += word
			} else {
				row[rowIndex] += " " + word
			}
		}

		// Ignore empty lines
		if len(row[0]) == 0 {
			continue
		}

		if lineID == 0 {
			// Draw the table head
			result += "<thead>"
			for i := 0; i < 6; i++ {
				result += "<th scope=\"col\">" + row[i] + "</th>"
			}
			result += "</thead><tbody>"
		} else {
			// Draw the row in red if the link isn't up
			result += "<tr class=\"" + (map[string]string{
				"up":      "table-success",
				"down":    "table-secondary",
				"start":   "table-danger",
				"passive": "table-info",
			})[row[3]] + "\">"
			// Add link to detail for first column
			if isIPv6 {
				result += "<td><a href=\"/ipv6/detail/" + serverName + "/" + row[0] + "\">" + row[0] + "</a></td>"
			} else {
				result += "<td><a href=\"/ipv4/detail/" + serverName + "/" + row[0] + "\">" + row[0] + "</a></td>"
			}
			// Draw the other cells
			for i := 1; i < 6; i++ {
				result += "<td>" + row[i] + "</td>"
			}
			result += "</tr>"
		}
	}
	result += "</tbody></table>"
	return result
}
