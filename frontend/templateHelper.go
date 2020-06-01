package main

import (
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Helper to check if the IP is valid
func isIP(s string) bool {
	return nil != net.ParseIP(s)
}

// Helper to check if the number is valid
func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return nil == err
}

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
		var isASes bool = false

		var lineFormatted string
		words := strings.Split(line, " ")

		for wordID, word := range words {
			if len(word) == 0 {
				continue
			}
			if wordID > 0 && (len(words[wordID-1]) == 0 || words[wordID-1][len(words[wordID-1])-1] == ':') {
				// Insert TAB if there are multiple spaces before this word
				lineFormatted += "\t"
			} else {
				lineFormatted += " "
			}

			if isIP(word) {
				// Add whois link to the IP, handles IPv4 and IPv6
				lineFormatted += "<a href=\"/whois/" + word + "\">" + word + "</a>"
			} else if len(strings.Split(word, "%")) == 2 && isIP(strings.Split(word, "%")[0]) {
				// IPv6 link-local with interface name, like fd00::1%eth0
				// Add whois link to address part
				lineFormatted += "<a href=\"/whois/" + strings.Split(word, "%")[0] + "\">" + strings.Split(word, "%")[0] + "</a>"
				lineFormatted += "%" + strings.Split(word, "%")[1]
			} else if len(strings.Split(word, "/")) == 2 && isIP(strings.Split(word, "/")[0]) {
				// IP with a CIDR range, like 192.168.0.1/24
				// Add whois link to first part
				lineFormatted += "<a href=\"/whois/" + strings.Split(word, "/")[0] + "\">" + strings.Split(word, "/")[0] + "</a>"
				lineFormatted += "/" + strings.Split(word, "/")[1]
			} else if word == "AS:" || word == "\tBGP.as_path:" {
				// Bird will output ASNs later
				isASes = true
				lineFormatted += word
			} else if isASes && isNumber(strings.Trim(word, "()")) {
				// Remove brackets in path caused by confederation
				wordNum := strings.Trim(word, "()")
				// Bird is outputing ASNs, add whois for them
				lineFormatted += "<a href=\"/whois/AS" + wordNum + "\">" + word + "</a>"
			} else {
				// Just an ordinary word, print it and done
				lineFormatted += word
			}
		}
		lineFormatted += "\n"
		result += lineFormatted
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
