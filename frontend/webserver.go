package main

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/handlers"
)

var primitiveMap = map[string]string{
	"summary":         "show protocols",
	"detail":          "show protocols all %s",
	"route":           "show route for %s",
	"route_all":       "show route for %s all",
	"route_where":     "show route where net ~ [ %s ]",
	"route_where_all": "show route where net ~ [ %s ] all",
	"route_generic":   "show route %s",
	"generic":         "show %s",
	"traceroute":      "%s",
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
		buffer.String(),
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
			tmp, err := url.PathUnescape(split[2])
			if err != nil {
				serverError(w, r)
				return
			}
			urlCommands = tmp
		}

		var backendCommand string
		if strings.Contains(backendCommandPrimitive, "%") {
			backendCommand = fmt.Sprintf(backendCommandPrimitive, urlCommands)
		} else {
			backendCommand = backendCommandPrimitive
		}
		backendCommand = strings.TrimSpace(backendCommand)

		escapedServers, err := url.PathUnescape(split[1])
		if err != nil {
			serverError(w, r)
			return
		}
		servers := strings.Split(escapedServers, "+")

		var responses []string = batchRequest(servers, endpoint, backendCommand)
		var content string
		for i, response := range responses {

			var result string
			if (endpoint == "bird") && backendCommand == "show protocols" && len(response) > 4 && strings.ToLower(response[0:4]) == "name" {
				result = summaryTable(response, servers[i])
			} else {
				result = smartFormatter(response)
			}

			// render the bird result template
			args := TemplateBird{
				ServerName: servers[i],
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
			content,
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

		// render the bgpmap result template
		args := TemplateBGPmap{
			Servers: servers,
			Target:  backendCommand,
			Result:  birdRouteToGraphviz(servers, responses, urlCommands),
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
			buffer.String(),
		)
	}
}

// redirect from the form input to a path style query
func webHandlerNavbarFormRedirect(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	action := query.Get("action")

	switch action {
	case "whois":
		target := url.PathEscape(query.Get("target"))
		http.Redirect(w, r, "/"+action+"/"+target, 302)
	case "summary":
		server := url.PathEscape(query.Get("server"))
		http.Redirect(w, r, "/"+action+"/"+server+"/", 302)
	default:
		server := url.PathEscape(query.Get("server"))
		target := url.PathEscape(query.Get("target"))
		http.Redirect(w, r, "/"+action+"/"+server+"/"+target, 302)
	}
}

// set up routing paths and start webserver
func webServerStart() {

	// redirect main page to all server summary
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/summary/"+strings.Join(setting.servers, "+"), 302)
	})

	// serve static pages using the AssetFS and bindata
	fs := http.FileServer(&assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
		Prefix:    "",
	})

	http.Handle("/static/", fs)
	http.Handle("/robots.txt", fs)
	http.Handle("/favicon.ico", fs)

	// backend routes
	http.HandleFunc("/summary/", webBackendCommunicator("bird", "summary"))
	http.HandleFunc("/detail/", webBackendCommunicator("bird", "detail"))
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
	http.HandleFunc("/redir", webHandlerNavbarFormRedirect)
	http.HandleFunc("/telegram/", webHandlerTelegramBot)

	// Start HTTP server
	http.ListenAndServe(setting.listen, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
}
