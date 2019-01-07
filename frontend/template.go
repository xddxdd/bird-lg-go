package main

import (
    "net"
    "net/http"
    "strings"
    "strconv"
)

func templateHeader(w http.ResponseWriter, r *http.Request, title string) {
    path := r.URL.Path
    if len(strings.Split(r.URL.Path, "/")) < 4 {
        path = "/ipv4/summary/" + strings.Join(settingServers[:], "+") + "/"
    }

    split := strings.Split(path, "/")
    split[1] = "ipv4"
    ipv4_url := strings.Join(split, "/")

    split = strings.Split(path, "/")
    split[1] = "ipv6"
    ipv6_url := strings.Join(split, "/")

    split = strings.Split(path, "/")
    split[3] = strings.Join(settingServers[:], "+")
    all_url := strings.Join(split, "/")

    split = strings.Split(path, "/")
    var serverAllActive string
    if split[3] == strings.Join(settingServers[:], "+") {
        serverAllActive = " active"
    }

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
    if split[2] == "whois" {
        options += `<option value="whois" selected>whois ...</option>`
    } else {
        options += `<option value="whois">whois ...</option>`
    }

    var target string
    if len(split) >= 5 {
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

func templateFooter(w http.ResponseWriter) {
    w.Write([]byte(`
</div>
<!--<script data-no-instant src="https://cdn.jsdelivr.net/npm/jquery@3.3.1/dist/jquery.min.js"></script>
<script data-no-instant src="https://cdn.jsdelivr.net/npm/bootstrap@4.2.1/dist/js/bootstrap.bundle.min.js"></script>-->
</body>
</html>
    `))
}

func isIP(s string) bool {
    return nil != net.ParseIP(s)
}

func isNumber(s string) bool {
    _, err := strconv.Atoi(s)
    return nil == err
}

func smartWriter(w http.ResponseWriter, s string) {
    w.Write([]byte("<pre>"))
    for _, line := range strings.Split(s, "\n") {
        var tabPending bool = false
        var isFirstWord bool = true
        var isASes bool = false
        for _, word := range strings.Split(line, " ") {
            if len(word) == 0 {
                tabPending = true
            } else {
                if isFirstWord {
                    isFirstWord = false
                } else if tabPending {
                    w.Write([]byte("\t"))
                    tabPending = false
                } else {
                    w.Write([]byte(" "))
                }

                if isIP(word) {
                    w.Write([]byte("<a href=\"/whois/" + word + "\">" + word + "</a>"))
                } else if len(strings.Split(word, "%")) == 2 && isIP(strings.Split(word, "%")[0]) {
                    w.Write([]byte("<a href=\"/whois/" + strings.Split(word, "%")[0] + "\">" + strings.Split(word, "%")[0] + "</a>"))
                    w.Write([]byte("%" + strings.Split(word, "%")[1]))
                } else if len(strings.Split(word, "/")) == 2 && isIP(strings.Split(word, "/")[0]) {
                    w.Write([]byte("<a href=\"/whois/" + strings.Split(word, "/")[0] + "\">" + strings.Split(word, "/")[0] + "</a>"))
                    w.Write([]byte("/" + strings.Split(word, "/")[1]))
                } else if word == "AS:" || word == "\tBGP.as_path:" {
                    isASes = true
                    w.Write([]byte(word))
                } else if isASes && isNumber(word) {
                    w.Write([]byte("<a href=\"/whois/AS" + word + "\">" + word + "</a>"))
                } else {
                    w.Write([]byte(word))
                }
            }
        }
        w.Write([]byte("\n"))
    }
    w.Write([]byte("</pre>"))
}

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

        if len(row[0]) == 0 {
            continue
        }
        if lineId == 0 {
            w.Write([]byte("<thead>"))
            for i := 0; i < 6; i++ {
                w.Write([]byte("<th scope=\"col\">" + row[i] + "</th>"))
            }
            w.Write([]byte("</thead><tbody>"))
        } else {
            if row[3] == "up" {
                w.Write([]byte("<tr>"))
            } else if lineId != 0 {
                w.Write([]byte("<tr class=\"table-danger\">"))
            }
            if isIPv6 {
                w.Write([]byte("<td><a href=\"/ipv6/detail/" + serverName + "/" +  row[0] + "\">" + row[0] + "</a></td>"))
            } else {
                w.Write([]byte("<td><a href=\"/ipv4/detail/" + serverName + "/" +  row[0] + "\">" + row[0] + "</a></td>"))
            }
            for i := 1; i < 6; i++ {
                w.Write([]byte("<td>" + row[i] + "</td>"))
            }
            w.Write([]byte("</tr>"))
        }
    }
    w.Write([]byte("</tbody></table>"))
}
