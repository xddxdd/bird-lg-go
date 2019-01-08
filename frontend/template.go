package main

import (
    "net"
    "net/http"
    "strings"
    "strconv"
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
    path := r.URL.Path
    split := strings.Split(r.URL.Path, "/")

    // Mark if the URL is for a whois query
    var isWhois bool = false
    if len(split) >= 2 && split[1] == "whois" {
        isWhois = true
    }

    // Use a default URL if the request URL is too short
    // The URL is for return to IPv4 summary page
    if len(split) < 4 {
        path = "/ipv4/summary/" + strings.Join(settingServers[:], "+") + "/"
    } else if len(split) == 4 {
        path += "/"
    }

    // Compose URLs for link in navbar
    split = strings.Split(path, "/")
    split[1] = "ipv4"
    ipv4_url := strings.Join(split, "/")

    split = strings.Split(path, "/")
    split[1] = "ipv6"
    ipv6_url := strings.Join(split, "/")

    split = strings.Split(path, "/")
    split[3] = strings.Join(settingServers[:], "+")
    all_url := strings.Join(split, "/")

    // Check if the "All Server" link should be marked as active
    split = strings.Split(path, "/")
    var serverAllActive string
    if split[3] == strings.Join(settingServers[:], "+") {
        serverAllActive = " active"
    }

    // Print the IPv4, IPv6, All Servers link in navbar
    var serverNavigation string = `
        <li class="nav-item">
            <a class="nav-link" href="` + ipv4_url + `"> IPv4 </a>
        </li>
        <li class="nav-item">
            <a class="nav-link" href="` + ipv6_url + `"> IPv6 </a>
        </li>
        <span class="navbar-text">|</span>
        <li class="nav-item">
            <a class="nav-link` + serverAllActive + `" href="` + all_url + `"> All Servers </a>
        </li>
    `

    // Add a link for each of the servers
    for _, server := range settingServers {
        split = strings.Split(path, "/")
        var serverActive string
        if split[3] == server {
            serverActive = " active"
        }
        split[3] = server
        server_url := strings.Join(split, "/")

        serverNavigation += `
            <li class="nav-item">
                <a class="nav-link` + serverActive + `" href="` + server_url + `">` + server + `</a>
            </li>
        `
    }

    // Add the options in navbar form, and check if they are active
    var options string
    split = strings.Split(path, "/")
    if split[2] == "summary" {
        options += `<option value="summary" selected>show protocol</option>`
    } else {
        options += `<option value="summary">show protocol</option>`
    }
    if split[2] == "route" {
        options += `<option value="route" selected>show route for ...</option>`
    } else {
        options += `<option value="route">show route for ...</option>`
    }
    if split[2] == "route_all" {
        options += `<option value="route_all" selected>show route for ... all</option>`
    } else {
        options += `<option value="route_all">show route for ... all</option>`
    }
    if isWhois {
        options += `<option value="whois" selected>whois ...</option>`
    } else {
        options += `<option value="whois">whois ...</option>`
    }
    if split[2] == "traceroute" {
        options += `<option value="traceroute" selected>traceroute ...</option>`
    } else {
        options += `<option value="traceroute">traceroute ...</option>`
    }

    var target string
    if isWhois {
        // This is a whois request, use original path URL instead of the modified one
        // and extract the target
        whoisSplit := strings.Split(r.URL.Path, "/")
        target = whoisSplit[2]
    } else if len(split) >= 5 {
        // This is a normal request, just extract the target
        target = split[4]
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
<link href="https://cdn.jsdelivr.net/npm/bootstrap@4.2.1/dist/css/bootstrap.min.css" rel="stylesheet">
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
                <input name="proto" class="d-none" value="` + split[1] + `">
                <input name="server" class="d-none" value="` + split[3] + `">
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
        var tabPending bool = false
        var isFirstWord bool = true
        var isASes bool = false
        for _, word := range strings.Split(line, " ") {
            // Process each word
            if len(word) == 0 {
                // Indicates that two spaces are connected together
                // Replace this with a tab later
                tabPending = true
            } else {
                if isFirstWord {
                    // Do not add space before the first word
                    isFirstWord = false
                } else if tabPending {
                    // A tab should be added; add it
                    w.Write([]byte("\t"))
                    tabPending = false
                } else {
                    // Two words separated by a space, just print the space
                    w.Write([]byte(" "))
                }

                if isIP(word) {
                    // Add whois link to the IP, handles IPv4 and IPv6
                    w.Write([]byte("<a href=\"/whois/" + word + "\">" + word + "</a>"))
                } else if len(strings.Split(word, "%")) == 2 && isIP(strings.Split(word, "%")[0]) {
                    // IPv6 link-local with interface name, like fd00::1%eth0
                    // Add whois link to address part
                    w.Write([]byte("<a href=\"/whois/" + strings.Split(word, "%")[0] + "\">" + strings.Split(word, "%")[0] + "</a>"))
                    w.Write([]byte("%" + strings.Split(word, "%")[1]))
                } else if len(strings.Split(word, "/")) == 2 && isIP(strings.Split(word, "/")[0]) {
                    // IP with a CIDR range, like 192.168.0.1/24
                    // Add whois link to first part
                    w.Write([]byte("<a href=\"/whois/" + strings.Split(word, "/")[0] + "\">" + strings.Split(word, "/")[0] + "</a>"))
                    w.Write([]byte("/" + strings.Split(word, "/")[1]))
                } else if word == "AS:" || word == "\tBGP.as_path:" {
                    // Bird will output ASNs later
                    isASes = true
                    w.Write([]byte(word))
                } else if isASes && isNumber(word) {
                    // Bird is outputing ASNs, ass whois for them
                    w.Write([]byte("<a href=\"/whois/AS" + word + "\">" + word + "</a>"))
                } else {
                    // Just an ordinary word, print it and done
                    w.Write([]byte(word))
                }
            }
        }
        w.Write([]byte("\n"))
    }
    w.Write([]byte("</pre>"))
}

// Output a table for the summary page
func summaryTable(w http.ResponseWriter, isIPv6 bool, data string, serverName string) {
    w.Write([]byte("<table class=\"table table-striped table-bordered table-sm\">"))
    for lineId, line := range strings.Split(data, "\n") {
        var tabPending bool = false
        var tableCells int = 0
        var row [6]string
        for i, word := range strings.Split(line, " ") {
            if len(word) == 0 {
                tabPending = true
            } else {
                if i == 0 {
                    tabPending = true
                } else if tabPending {
                    // Allow up to 6 columns in the table, any more is ignored
                    if tableCells < 5 {
                        tableCells++
                    } else {
                        row[tableCells] += " "
                    }
                    tabPending = false
                } else {
                    row[tableCells] += " "
                }
                row[tableCells] += word
            }
        }

        // Ignore empty lines
        if len(row[0]) == 0 {
            continue
        }

        if lineId == 0 {
            // Draw the table head
            w.Write([]byte("<thead>"))
            for i := 0; i < 6; i++ {
                w.Write([]byte("<th scope=\"col\">" + row[i] + "</th>"))
            }
            w.Write([]byte("</thead><tbody>"))
        } else {
            // Draw the row in red if the link isn't up
            if row[3] == "up" {
                w.Write([]byte("<tr>"))
            } else if lineId != 0 {
                w.Write([]byte("<tr class=\"table-danger\">"))
            }
            // Add link to detail for first column
            if isIPv6 {
                w.Write([]byte("<td><a href=\"/ipv6/detail/" + serverName + "/" +  row[0] + "\">" + row[0] + "</a></td>"))
            } else {
                w.Write([]byte("<td><a href=\"/ipv4/detail/" + serverName + "/" +  row[0] + "\">" + row[0] + "</a></td>"))
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
