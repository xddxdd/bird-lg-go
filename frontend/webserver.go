package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

var primitiveMap = map[string]string{
	"summary":                          "show protocols",
	"detail":                           "show protocols all %s",
	"route_from_protocol":              "show route protocol %s",
	"route_from_protocol_all":          "show route protocol %s all",
	"route_from_protocol_primary":      "show route protocol %s primary",
	"route_from_protocol_all_primary":  "show route protocol %s all primary",
	"route_filtered_from_protocol":     "show route filtered protocol %s",
	"route_filtered_from_protocol_all": "show route filtered protocol %s all",
	"route_from_origin":                "show route where bgp_path.last = %s",
	"route_from_origin_all":            "show route where bgp_path.last = %s all",
	"route_from_origin_primary":        "show route where bgp_path.last = %s primary",
	"route_from_origin_all_primary":    "show route where bgp_path.last = %s all primary",
	"route":                            "show route for %s",
	"route_all":                        "show route for %s all",
	"route_where":                      "show route where net ~ [ %s ]",
	"route_where_all":                  "show route where net ~ [ %s ] all",
	"route_generic":                    "show route %s",
	"generic":                          "show %s",
	"whois":                            "%s",
	"traceroute":                       "%s",
}

// serve up a generic error
func serverError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

// WHOIS pages
func webHandlerWhois(w http.ResponseWriter, r *http.Request) {
	target, err := url.PathUnescape(r.URL.Path[len("/whois/"):])
	if err != nil {
		serverError(w, r)
		return
	}

	// render the whois template
	args := TemplateWhois{
		Target: target,
		Result: smartFormatter(whois(target)),
	}

	tmpl := TemplateLibrary["whois"]
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, args)
	if err != nil {
		fmt.Println("Error rendering whois template:", err.Error())
	}

	renderPageTemplate(
		w, r,
		" - whois "+html.EscapeString(target),
		template.HTML(buffer.String()),
	)
}

// serve up results from bird
func webBackendCommunicator(endpoint string, command string) func(w http.ResponseWriter, r *http.Request) {

	backendCommandPrimitive, commandPresent := primitiveMap[command]
	if !commandPresent {
		panic("invalid command: " + command)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		split := strings.SplitN(r.URL.Path[1:], "/", 3)
		var urlCommands string
		if len(split) >= 3 {
			urlCommands = split[2]
		}

		var backendCommand string
		if strings.Contains(backendCommandPrimitive, "%") {
			backendCommand = fmt.Sprintf(backendCommandPrimitive, urlCommands)
		} else {
			backendCommand = backendCommandPrimitive
		}
		backendCommand = strings.TrimSpace(backendCommand)

		servers := strings.Split(split[1], "+")

		var responses []string = batchRequest(servers, endpoint, backendCommand)
		var content string
		for i, response := range responses {

			var result template.HTML
			if (endpoint == "bird") && backendCommand == "show protocols" && len(response) > 4 && strings.ToLower(response[0:4]) == "name" {
				result = summaryTable(response, servers[i])
			} else {
				result = smartFormatter(response)
			}

			serverDisplay := servers[i]
			for k, v := range setting.servers {
				if servers[i] == v {
					serverDisplay = setting.serversDisplay[k]
					break
				}
			}

			// render the bird result template
			args := TemplateBird{
				ServerName: serverDisplay,
				Target:     backendCommand,
				Result:     result,
			}

			tmpl := TemplateLibrary["bird"]
			var buffer bytes.Buffer
			err := tmpl.Execute(&buffer, args)
			if err != nil {
				fmt.Println("Error rendering bird template:", err.Error())
			}

			content += buffer.String()
		}

		renderPageTemplate(
			w, r,
			" - "+endpoint+" "+backendCommand,
			template.HTML(content),
		)
	}
}

// bgpmap result
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
		urlCommands := strings.Join(split[2:], "/")

		var backendCommand string
		if strings.Contains(backendCommandPrimitive, "%") {
			backendCommand = fmt.Sprintf(backendCommandPrimitive, urlCommands)
		} else {
			backendCommand = backendCommandPrimitive
		}

		var servers []string = strings.Split(split[1], "+")
		var responses []string = batchRequest(servers, endpoint, backendCommand)

		// encode result with base64 to prevent xss
		result := birdRouteToGraphviz(servers, responses, urlCommands)
		result = base64.StdEncoding.EncodeToString([]byte(result))

		// render the bgpmap result template
		args := TemplateBGPmap{
			Servers: servers,
			Target:  backendCommand,
			Result:  result,
		}

		tmpl := TemplateLibrary["bgpmap"]
		var buffer bytes.Buffer
		err := tmpl.Execute(&buffer, args)
		if err != nil {
			fmt.Println("Error rendering bgpmap template:", err.Error())
		}

		renderPageTemplate(
			w, r,
			" - "+html.EscapeString(endpoint+" "+backendCommand),
			template.HTML(buffer.String()),
		)
	}
}

// set up routing paths and start webserver
func webServerStart(l net.Listener) {

	// redirect main page to all server summary
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/summary/"+url.PathEscape(strings.Join(setting.servers, "+")), 302)
	})

	// serve static pages using embedded assets from template.go
	subfs, err := fs.Sub(assets, "assets")
	if err != nil {
		panic("Webserver fs.sub failed: " + err.Error())
	}
	fs := http.FileServer(http.FS(subfs))

	http.Handle("/static/", fs)
	http.Handle("/robots.txt", fs)
	http.Handle("/favicon.ico", fs)

	// backend routes
	http.HandleFunc("/summary/", webBackendCommunicator("bird", "summary"))
	http.HandleFunc("/detail/", webBackendCommunicator("bird", "detail"))
	http.HandleFunc("/route_filtered_from_protocol/", webBackendCommunicator("bird", "route_filtered_from_protocol"))
	http.HandleFunc("/route_filtered_from_protocol_all/", webBackendCommunicator("bird", "route_filtered_from_protocol_all"))
	http.HandleFunc("/route_from_protocol/", webBackendCommunicator("bird", "route_from_protocol"))
	http.HandleFunc("/route_from_protocol_all/", webBackendCommunicator("bird", "route_from_protocol_all"))
	http.HandleFunc("/route_from_protocol_primary/", webBackendCommunicator("bird", "route_from_protocol_primary"))
	http.HandleFunc("/route_from_protocol_all_primary/", webBackendCommunicator("bird", "route_from_protocol_all_primary"))
	http.HandleFunc("/route_from_origin/", webBackendCommunicator("bird", "route_from_origin"))
	http.HandleFunc("/route_from_origin_all/", webBackendCommunicator("bird", "route_from_origin_all"))
	http.HandleFunc("/route_from_origin_primary/", webBackendCommunicator("bird", "route_from_origin_primary"))
	http.HandleFunc("/route_from_origin_all_primary/", webBackendCommunicator("bird", "route_from_origin_all_primary"))
	http.HandleFunc("/route/", webBackendCommunicator("bird", "route"))
	http.HandleFunc("/route_all/", webBackendCommunicator("bird", "route_all"))
	http.HandleFunc("/route_bgpmap/", webHandlerBGPMap("bird", "route_bgpmap"))
	http.HandleFunc("/route_where/", webBackendCommunicator("bird", "route_where"))
	http.HandleFunc("/route_where_all/", webBackendCommunicator("bird", "route_where_all"))
	http.HandleFunc("/route_where_bgpmap/", webHandlerBGPMap("bird", "route_where_bgpmap"))
	http.HandleFunc("/route_generic/", webBackendCommunicator("bird", "route_generic"))
	http.HandleFunc("/generic/", webBackendCommunicator("bird", "generic"))
	http.HandleFunc("/traceroute/", webBackendCommunicator("traceroute", "traceroute"))
	http.HandleFunc("/whois/", webHandlerWhois)
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/telegram/", webHandlerTelegramBot)

	// Start HTTP server
	http.Serve(l, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
}
