package main

import (
	"html"
	"net/http"
	"strings"
)

func webHandlerWhois(w http.ResponseWriter, r *http.Request) {
	var target string = r.URL.Path[len("/whois/"):]

	templateHeader(w, r, "Bird-lg Go - whois "+html.EscapeString(target))

	w.Write([]byte("<h2>whois " + html.EscapeString(target) + "</h2>"))
	smartWriter(w, whois(target))

	templateFooter(w)
}

func webBackendCommunicator(w http.ResponseWriter, r *http.Request, endpoint string, command string) {
	split := strings.Split(r.URL.Path[1:], "/")
	urlCommands := strings.Join(split[3:], "/")

	command = (map[string]string{
		"summary":         "show protocols",
		"detail":          "show protocols all " + urlCommands,
		"route":           "show route for " + urlCommands,
		"route_all":       "show route for " + urlCommands + " all",
		"route_where":     "show route where net ~ [ " + urlCommands + " ]",
		"route_where_all": "show route where net ~ [ " + urlCommands + " ] all",
		"traceroute":      urlCommands,
	})[command]

	templateHeader(w, r, "Bird-lg Go - "+html.EscapeString(endpoint+" "+command))

	var servers []string = strings.Split(split[2], "+")

	var responses []string = batchRequest(servers, endpoint, command)
	for i, response := range responses {
		w.Write([]byte("<h2>" + html.EscapeString(servers[i]) + ": " + html.EscapeString(command) + "</h2>"))
		if (endpoint == "bird" || endpoint == "bird6") && command == "show protocols" && len(response) > 4 && strings.ToLower(response[0:4]) == "name" {
			var isIPv6 bool = endpoint[len(endpoint)-1] == '6'
			summaryTable(w, isIPv6, response, servers[i])
		} else {
			smartWriter(w, response)
		}
	}

	templateFooter(w)
}

func webHandlerNavbarFormRedirect(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query.Get("action") == "whois" {
		http.Redirect(w, r, "/"+query.Get("action")+"/"+query.Get("target"), 302)
	} else if query.Get("action") == "summary" {
		http.Redirect(w, r, "/"+query.Get("proto")+"/"+query.Get("action")+"/"+query.Get("server"), 302)
	} else {
		http.Redirect(w, r, "/"+query.Get("proto")+"/"+query.Get("action")+"/"+query.Get("server")+"/"+query.Get("target"), 302)
	}
}

func webServerStart() {
	// Start HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ipv4/summary/"+strings.Join(settingServers[:], "+"), 302)
	})
	http.HandleFunc("/ipv4/summary/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird", "summary") })
	http.HandleFunc("/ipv6/summary/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird6", "summary") })
	http.HandleFunc("/ipv4/detail/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird", "detail") })
	http.HandleFunc("/ipv6/detail/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird6", "detail") })
	http.HandleFunc("/ipv4/route/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird", "route") })
	http.HandleFunc("/ipv6/route/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird6", "route") })
	http.HandleFunc("/ipv4/route_all/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird", "route_all") })
	http.HandleFunc("/ipv6/route_all/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird6", "route_all") })
	http.HandleFunc("/ipv4/route_where/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird", "route_where") })
	http.HandleFunc("/ipv6/route_where/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird6", "route_where") })
	http.HandleFunc("/ipv4/route_where_all/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird", "route_where_all") })
	http.HandleFunc("/ipv6/route_where_all/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "bird6", "route_where_all") })
	http.HandleFunc("/ipv4/traceroute/", func(w http.ResponseWriter, r *http.Request) { webBackendCommunicator(w, r, "traceroute", "traceroute") })
	http.HandleFunc("/ipv6/traceroute/", func(w http.ResponseWriter, r *http.Request) {
		webBackendCommunicator(w, r, "traceroute6", "traceroute")
	})
	http.HandleFunc("/whois/", webHandlerWhois)
	http.HandleFunc("/redir/", webHandlerNavbarFormRedirect)
	http.ListenAndServe(settingListen, nil)
}
