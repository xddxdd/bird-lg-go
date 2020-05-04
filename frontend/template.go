package main

import (
	"net"
	"net/http"
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

// Print HTML header to the given http response
func templateHeader(w http.ResponseWriter, r *http.Request, title string) {
	path := r.URL.Path[1:]
	split := strings.Split(path, "/")

	// Mark if the URL is for a whois query
	var isWhois bool = split[0] == "whois"
	var whoisTarget string = strings.Join(split[1:], "/")

	// Use a default URL if the request URL is too short
	// The URL is for return to IPv4 summary page
	if len(split) < 3 {
		path = "ipv4/summary/" + strings.Join(setting.servers, "+") + "/"
	} else if len(split) == 3 {
		path += "/"
	}

	split = strings.Split(path, "/")

	// Compose URLs for link in navbar
	ipv4URL := "/" + strings.Join([]string{"ipv4", split[1], split[2], strings.Join(split[3:], "/")}, "/")
	ipv6URL := "/" + strings.Join([]string{"ipv6", split[1], split[2], strings.Join(split[3:], "/")}, "/")
	allURL := "/" + strings.Join([]string{split[0], split[1], strings.Join(setting.servers, "+"), strings.Join(split[3:], "/")}, "/")

	// Check if the "All Server" link should be marked as active
	var serverAllActive bool = strings.ToLower(split[2]) == strings.ToLower(strings.Join(setting.servers, "+"))

	// Print the IPv4, IPv6, All Servers link in navbar
	var serverNavigation string = `
        <li class="nav-item"><a class="nav-link` + (map[bool]string{true: " active"})[strings.ToLower(split[0]) == "ipv4"] + `" href="` + ipv4URL + `"> IPv4 </a></li>
        <li class="nav-item"><a class="nav-link` + (map[bool]string{true: " active"})[strings.ToLower(split[0]) == "ipv6"] + `" href="` + ipv6URL + `"> IPv6 </a></li>
        <span class="navbar-text">|</span>
        <li class="nav-item">
            <a class="nav-link` + (map[bool]string{true: " active"})[serverAllActive] + `" href="` + allURL + `"> All Servers </a>
        </li>`

	// Add a link for each of the servers
	for _, server := range setting.servers {
		var serverActive string
		if split[2] == server {
			serverActive = " active"
		}
		serverURL := "/" + strings.Join([]string{split[0], split[1], server, strings.Join(split[3:], "/")}, "/")

		serverNavigation += `
            <li class="nav-item">
                <a class="nav-link` + serverActive + `" href="` + serverURL + `">` + server + `</a>
            </li>`
	}

	// Add the options in navbar form, and check if they are active
	var optionKeys = []string{
		"summary",
		"detail",
		"route",
		"route_all",
		"route_bgpmap",
		"route_where",
		"route_where_all",
		"route_where_bgpmap",
		"whois",
		"traceroute",
	}
	var optionDisplays = []string{
		"show protocol",
		"show protocol all",
		"show route for ...",
		"show route for ... all",
		"show route for ... (bgpmap)",
		"show route where net ~ [ ... ]",
		"show route where net ~ [ ... ] all",
		"show route where net ~ [ ... ] (bgpmap)",
		"whois ...",
		"traceroute ...",
	}

	var options string
	for optionKeyID, optionKey := range optionKeys {
		options += "<option value=\"" + optionKey + "\""
		if (optionKey == "whois" && isWhois) || optionKey == split[1] {
			options += " selected"
		}
		options += ">" + optionDisplays[optionKeyID] + "</option>"
	}

	var target string
	if isWhois {
		// This is a whois request, use original path URL instead of the modified one
		// and extract the target
		target = whoisTarget
	} else if len(split) >= 4 {
		// This is a normal request, just extract the target
		target = strings.Join(split[3:], "/")
	}

	w.Write([]byte(`
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="zh-CN" lang="zh-CN" class="no-js">
<head>
<meta http-equiv="Content-Type" content="text/html;charset=UTF-8" />
<meta http-equiv="X-UA-Compatible" content="IE=edge"/>
<meta name="viewport" content="width=device-width,initial-scale=1,shrink-to-fit=no"/>
<meta name="renderer" content="webkit"/>
<title>` + title + `</title>
<link href="https://cdn.jsdelivr.net/npm/bootstrap@4.4.1/dist/css/bootstrap.min.css" rel="stylesheet">
<script src="https://cdn.jsdelivr.net/npm/viz.js@2.1.2/viz.min.js" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/viz.js@2.1.2/lite.render.js" crossorigin="anonymous"></script>
</head>
<body>

<nav class="navbar navbar-expand-lg navbar-light bg-light">
    <a class="navbar-brand" href="/">Bird-lg Go</a>
    <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
    </button>

    <div class="collapse navbar-collapse" id="navbarSupportedContent">
        <ul class="navbar-nav mr-auto">` + serverNavigation + `</ul>
        <form class="form-inline" action="/redir" method="GET">
            <div class="input-group">
                <select name="action" class="form-control">` + options + `</select>
                <input name="proto" class="d-none" value="` + split[0] + `">
                <input name="server" class="d-none" value="` + split[2] + `">
                <input name="target" class="form-control" placeholder="Target" aria-label="Target" value="` + target + `">
                <div class="input-group-append">
                    <button class="btn btn-outline-success" type="submit">&raquo;</button>
                </div>
            </div>
        </form>
    </div>
</nav>

<div class="container">
    `))
}

// Print HTML footer to http response
func templateFooter(w http.ResponseWriter) {
	w.Write([]byte(`
</div>
</body>
</html>
    `))
}

// Write the given text to http response, and add whois links for
// ASNs and IP addresses
func smartWriter(w http.ResponseWriter, s string) {
	w.Write([]byte("<pre>"))
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
		w.Write([]byte(lineFormatted))
	}
	w.Write([]byte("</pre>"))
}

// Output a table for the summary page
func summaryTable(w http.ResponseWriter, isIPv6 bool, data string, serverName string) {
	// w.Write([]byte("<pre>" + data + "</pre>"))
	w.Write([]byte("<table class=\"table table-striped table-bordered table-sm\">"))
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
			w.Write([]byte("<thead>"))
			for i := 0; i < 6; i++ {
				w.Write([]byte("<th scope=\"col\">" + row[i] + "</th>"))
			}
			w.Write([]byte("</thead><tbody>"))
		} else {
			// Draw the row in red if the link isn't up
			w.Write([]byte("<tr class=\"" + (map[string]string{
				"up":      "table-success",
				"down":    "table-secondary",
				"start":   "table-danger",
				"passive": "table-info",
			})[row[3]] + "\">"))
			// Add link to detail for first column
			if isIPv6 {
				w.Write([]byte("<td><a href=\"/ipv6/detail/" + serverName + "/" + row[0] + "\">" + row[0] + "</a></td>"))
			} else {
				w.Write([]byte("<td><a href=\"/ipv4/detail/" + serverName + "/" + row[0] + "\">" + row[0] + "</a></td>"))
			}
			// Draw the other cells
			for i := 1; i < 6; i++ {
				w.Write([]byte("<td>" + row[i] + "</td>"))
			}
			w.Write([]byte("</tr>"))
		}
	}
	w.Write([]byte("</tbody></table>"))
}
