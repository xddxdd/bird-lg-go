package main

import (
    "net/http"
    "strings"
    "html"
)

func webDispatcherIPv4Summary(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv4/summary/"):], "/")
    webHandler(w, r, "bird", split[0], "show protocols")
}

func webDispatcherIPv6Summary(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv6/summary/"):], "/")
    webHandler(w, r, "bird6", split[0], "show protocols")
}

func webDispatcherIPv4Detail(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv4/detail/"):], "/")
    webHandler(w, r, "bird", split[0], "show protocols all " + split[1])
}

func webDispatcherIPv6Detail(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv6/detail/"):], "/")
    webHandler(w, r, "bird6", split[0], "show protocols all " + split[1])
}

func webDispatcherIPv4Route(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv4/route/"):], "/")
    webHandler(w, r, "bird", split[0], "show route for " + strings.Join(split[1:], "/"))
}

func webDispatcherIPv6Route(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv6/route/"):], "/")
    webHandler(w, r, "bird6", split[0], "show route for " + strings.Join(split[1:], "/"))
}

func webDispatcherIPv4RouteAll(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv4/route_all/"):], "/")
    webHandler(w, r, "bird", split[0], "show route for " + strings.Join(split[1:], "/") + " all")
}

func webDispatcherIPv6RouteAll(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv6/route_all/"):], "/")
    webHandler(w, r, "bird6", split[0], "show route for " + strings.Join(split[1:], "/") + " all")
}

func webDispatcherIPv4RouteWhere(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv4/route_where/"):], "/")
    webHandler(w, r, "bird", split[0], "show route where net ~ [ " + strings.Join(split[1:], "/") + " ]")
}

func webDispatcherIPv6RouteWhere(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv6/route_where/"):], "/")
    webHandler(w, r, "bird6", split[0], "show route where net ~ [ " + strings.Join(split[1:], "/") + " ]")
}

func webDispatcherIPv4RouteWhereAll(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv4/route_where_all/"):], "/")
    webHandler(w, r, "bird", split[0], "show route where net ~ [ " + strings.Join(split[1:], "/") + " ] all")
}

func webDispatcherIPv6RouteWhereAll(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv6/route_where_all/"):], "/")
    webHandler(w, r, "bird6", split[0], "show route where net ~ [ " + strings.Join(split[1:], "/") + " ] all")
}

func webDispatcherWhois(w http.ResponseWriter, r *http.Request) {
    var target string = r.URL.Path[len("/whois/"):]

    templateHeader(w, r, "Bird-lg Go - whois " + html.EscapeString(target))

    w.Write([]byte("<h2>whois " + html.EscapeString(target) + "</h2>"))
    smartWriter(w, whois(target))

    templateFooter(w)
}

func webDispatcherIPv4Traceroute(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv4/traceroute/"):], "/")
    webHandler(w, r, "traceroute", split[0], strings.Join(split[1:], "/"))
}

func webDispatcherIPv6Traceroute(w http.ResponseWriter, r *http.Request) {
    split := strings.Split(r.URL.Path[len("/ipv6/traceroute/"):], "/")
    webHandler(w, r, "traceroute6", split[0], strings.Join(split[1:], "/"))
}

func webHandler(w http.ResponseWriter, r *http.Request, endpoint string, serverQuery string, command string) {
    templateHeader(w, r, "Bird-lg Go - " + html.EscapeString(endpoint + " " + command))

    var servers []string = strings.Split(serverQuery, "+")

    var responses []string = batchRequest(servers, endpoint, command)
    for i, response := range responses {
        w.Write([]byte("<h2>" + html.EscapeString(servers[i]) + ": " + html.EscapeString(command) + "</h2>"))
        if (endpoint == "bird" || endpoint == "bird6") && command == "show protocols" && strings.ToLower(response[0:4]) == "name" {
            var isIPv6 bool = endpoint[len(endpoint) - 1] == '6'
            summaryTable(w, isIPv6, response, servers[i])
        } else {
            smartWriter(w, response)
        }
    }

    templateFooter(w)
}

func defaultRedirect(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/ipv4/summary/" + strings.Join(settingServers[:], "+"), 302)
}

func navbarFormRedirect(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    if query.Get("action") == "whois" {
        http.Redirect(w, r, "/" + query.Get("action") + "/" + query.Get("target"), 302)
    } else if query.Get("action") == "summary" {
        http.Redirect(w, r, "/" + query.Get("proto") + "/" + query.Get("action") + "/" + query.Get("server"), 302)
    } else {
        http.Redirect(w, r, "/" + query.Get("proto") + "/" + query.Get("action") + "/" + query.Get("server") + "/" + query.Get("target"), 302)
    }
}

func webServerStart() {
    // Start HTTP server
    http.HandleFunc("/", defaultRedirect)
    http.HandleFunc("/ipv4/summary/", webDispatcherIPv4Summary)
    http.HandleFunc("/ipv6/summary/", webDispatcherIPv6Summary)
    http.HandleFunc("/ipv4/detail/", webDispatcherIPv4Detail)
    http.HandleFunc("/ipv6/detail/", webDispatcherIPv6Detail)
    http.HandleFunc("/ipv4/route/", webDispatcherIPv4Route)
    http.HandleFunc("/ipv6/route/", webDispatcherIPv6Route)
    http.HandleFunc("/ipv4/route_all/", webDispatcherIPv4RouteAll)
    http.HandleFunc("/ipv6/route_all/", webDispatcherIPv6RouteAll)
    http.HandleFunc("/ipv4/route_where/", webDispatcherIPv4RouteWhere)
    http.HandleFunc("/ipv6/route_where/", webDispatcherIPv6RouteWhere)
    http.HandleFunc("/ipv4/route_where_all/", webDispatcherIPv4RouteWhereAll)
    http.HandleFunc("/ipv6/route_where_all/", webDispatcherIPv6RouteWhereAll)
    http.HandleFunc("/ipv4/traceroute/", webDispatcherIPv4Traceroute)
    http.HandleFunc("/ipv6/traceroute/", webDispatcherIPv6Traceroute)
    http.HandleFunc("/whois/", webDispatcherWhois)
    http.HandleFunc("/redir/", navbarFormRedirect)
    http.ListenAndServe(settingListen, nil)
}
