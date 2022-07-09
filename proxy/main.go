package main

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

// Check if a byte is character for number
func isNumeric(b byte) bool {
	return b >= byte('0') && b <= byte('9')
}

// Default handler, returns 500 Internal Server Error
func invalidHandler(httpW http.ResponseWriter, httpR *http.Request) {
	httpW.WriteHeader(http.StatusInternalServerError)
	httpW.Write([]byte("Invalid Request\n"))
}

// Access handler, check to see if client IP in allowed IPs, continue if it is, send to invalidHandler if not
func accessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpW http.ResponseWriter, httpR *http.Request) {

		// setting.allowedIPs will always have at least one element because of how it's defined
		if setting.allowedIPs[0] == "" {
			next.ServeHTTP(httpW, httpR)
			return
		}

		IPPort := httpR.RemoteAddr

		// Remove port from IP and remove brackets that are around IPv6 addresses
		requestIp := IPPort[0:strings.LastIndex(IPPort, ":")]
		requestIp = strings.Replace(requestIp, "[", "", -1)
		requestIp = strings.Replace(requestIp, "]", "", -1)

		for _, allowedIP := range setting.allowedIPs {
			if requestIp == allowedIP {
				next.ServeHTTP(httpW, httpR)
				return
			}
		}

		invalidHandler(httpW, httpR)
		return
	})
}

type settingType struct {
	birdSocket string
	listen     string
	allowedIPs []string
	tr_bin     string
	tr_raw     bool
}

var setting settingType

// Wrapper of tracer
func main() {
	parseSettings()

	var l net.Listener
	var err error

	if strings.HasPrefix(setting.listen, "/") {
		// Delete existing socket file, ignore errors (will fail later anyway)
		os.Remove(setting.listen)
		l, err = net.Listen("unix", setting.listen)
	} else {
		listenAddr := setting.listen
		if !strings.Contains(listenAddr, ":") {
			listenAddr = ":" + listenAddr
		}
		l, err = net.Listen("tcp", listenAddr)
	}

	if err != nil {
		panic(err)
	}

	// Start HTTP server
	http.HandleFunc("/", invalidHandler)
	http.HandleFunc("/bird", birdHandler)
	http.HandleFunc("/bird6", birdHandler)
	http.HandleFunc("/traceroute", tracerouteHandler)
	http.HandleFunc("/traceroute6", tracerouteHandler)
	http.Serve(l, handlers.LoggingHandler(os.Stdout, accessHandler(http.DefaultServeMux)))
}
