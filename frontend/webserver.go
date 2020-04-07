package main

import (
	"fmt"
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

func webBackendCommunicator(endpoint string, command string) func(w http.ResponseWriter, r *http.Request) {
	backendCommandPrimitive, commandPresent := (map[string]string{
		"summary":         "show protocols",
		"detail":          "show protocols all %s",
		"route":           "show route for %s",
		"route_all":       "show route for %s all",
		"route_where":     "show route where net ~ [ %s ]",
		"route_where_all": "show route where net ~ [ %s ] all",
		"traceroute":      "%s",
	})[command]

	if !commandPresent {
		panic("invalid command: " + command)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		split := strings.Split(r.URL.Path[1:], "/")
		urlCommands := strings.Join(split[3:], "/")

		var backendCommand string
		if strings.Contains(backendCommandPrimitive, "%") {
			backendCommand = fmt.Sprintf(backendCommandPrimitive, urlCommands)
		} else {
			backendCommand = backendCommandPrimitive
		}

		templateHeader(w, r, "Bird-lg Go - "+html.EscapeString(endpoint+" "+backendCommand))

		var servers []string = strings.Split(split[2], "+")

		var responses []string = batchRequest(servers, endpoint, backendCommand)
		for i, response := range responses {
			w.Write([]byte("<h2>" + html.EscapeString(servers[i]) + ": " + html.EscapeString(backendCommand) + "</h2>"))
			if (endpoint == "bird" || endpoint == "bird6") && backendCommand == "show protocols" && len(response) > 4 && strings.ToLower(response[0:4]) == "name" {
				var isIPv6 bool = endpoint[len(endpoint)-1] == '6'
				summaryTable(w, isIPv6, response, servers[i])
			} else {
				smartWriter(w, response)
			}
		}

		templateFooter(w)
	}
}

func webHandlerBGPMap(endpoint string, command string) func(w http.ResponseWriter, r *http.Request) {
	backendCommandPrimitive, commandPresent := (map[string]string{
		"route_bgpmap":       "show route for %s all",
		"route_where_bgpmap": "show route where net ~ [ %s ] all",
	})[command]

	if !commandPresent {
		panic("invalid command: " + command)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		split := strings.Split(r.URL.Path[1:], "/")
		urlCommands := strings.Join(split[3:], "/")

		var backendCommand string
		if strings.Contains(backendCommandPrimitive, "%") {
			backendCommand = fmt.Sprintf(backendCommandPrimitive, urlCommands)
		} else {
			backendCommand = backendCommandPrimitive
		}

		templateHeader(w, r, "Bird-lg Go - "+html.EscapeString(endpoint+" "+backendCommand))

		var servers []string = strings.Split(split[2], "+")

		var responses []string = batchRequest(servers, endpoint, backendCommand)
		w.Write([]byte(`
		<script>
		var viz = new Viz();
		viz.renderSVGElement(` + "`" + birdRouteToGraphviz(servers, responses, urlCommands) + "`" + `)
		.then(function(element) {
			document.body.appendChild(element);
		})
		.catch(error => {
			document.body.appendChild("<pre>"+error+"</pre>")
		});
		</script>`))

		templateFooter(w)
	}
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
		http.Redirect(w, r, "/ipv4/summary/"+strings.Join(setting.servers, "+"), 302)
	})
	http.HandleFunc("/ipv4/summary/", webBackendCommunicator("bird", "summary"))
	http.HandleFunc("/ipv6/summary/", webBackendCommunicator("bird6", "summary"))
	http.HandleFunc("/ipv4/detail/", webBackendCommunicator("bird", "detail"))
	http.HandleFunc("/ipv6/detail/", webBackendCommunicator("bird6", "detail"))
	http.HandleFunc("/ipv4/route/", webBackendCommunicator("bird", "route"))
	http.HandleFunc("/ipv6/route/", webBackendCommunicator("bird6", "route"))
	http.HandleFunc("/ipv4/route_all/", webBackendCommunicator("bird", "route_all"))
	http.HandleFunc("/ipv6/route_all/", webBackendCommunicator("bird6", "route_all"))
	http.HandleFunc("/ipv4/route_bgpmap/", webHandlerBGPMap("bird", "route_bgpmap"))
	http.HandleFunc("/ipv6/route_bgpmap/", webHandlerBGPMap("bird6", "route_bgpmap"))
	http.HandleFunc("/ipv4/route_where/", webBackendCommunicator("bird", "route_where"))
	http.HandleFunc("/ipv6/route_where/", webBackendCommunicator("bird6", "route_where"))
	http.HandleFunc("/ipv4/route_where_all/", webBackendCommunicator("bird", "route_where_all"))
	http.HandleFunc("/ipv6/route_where_all/", webBackendCommunicator("bird6", "route_where_all"))
	http.HandleFunc("/ipv4/route_where_bgpmap/", webHandlerBGPMap("bird", "route_where_bgpmap"))
	http.HandleFunc("/ipv6/route_where_bgpmap/", webHandlerBGPMap("bird6", "route_where_bgpmap"))
	http.HandleFunc("/ipv4/traceroute/", webBackendCommunicator("traceroute", "traceroute"))
	http.HandleFunc("/ipv6/traceroute/", webBackendCommunicator("traceroute6", "traceroute"))
	http.HandleFunc("/whois/", webHandlerWhois)
	http.HandleFunc("/redir/", webHandlerNavbarFormRedirect)
	http.HandleFunc("/telegram/", webHandlerTelegramBot)
	http.ListenAndServe(setting.listen, nil)
}
