package main

import (
	"fmt"
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

func hasAccess(remoteAddr string) bool {
	// setting.allowedNets will always have at least one element because of how it's defined
	if len(setting.allowedNets) == 0 {
		return true
	}

	if !strings.Contains(remoteAddr, ":") {
		return false
	}

	// Remove port from IP and remove brackets that are around IPv6 addresses
	remoteAddr = remoteAddr[0:strings.LastIndex(remoteAddr, ":")]
	remoteAddr = strings.Trim(remoteAddr, "[]")

	ipObject := net.ParseIP(remoteAddr)
	if ipObject == nil {
		return false
	}

	for _, net := range setting.allowedNets {
		if net.Contains(ipObject) {
			return true
		}
	}

	return false
}

// Access handler, check to see if client IP in allowed nets, continue if it is, send to invalidHandler if not
func accessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpW http.ResponseWriter, httpR *http.Request) {
		if hasAccess(httpR.RemoteAddr) {
			next.ServeHTTP(httpW, httpR)
		} else {
			invalidHandler(httpW, httpR)
		}
	})
}

type settingType struct {
	birdSocket        string
	listen            []string
	allowedNets       []*net.IPNet
	tr_bin            string
	tr_flags          []string
	tr_raw            bool
	tr_max_concurrent int
}

var setting settingType

// Wrapper of tracer
func main() {
	parseSettings()
	initTracerouteSemaphore(setting.tr_max_concurrent)
	tracerouteAutodetect()

	mux := http.NewServeMux()

	// Prepare HTTP server
	mux.HandleFunc("/", invalidHandler)
	mux.HandleFunc("/bird", birdHandler)
	mux.HandleFunc("/bird6", birdHandler)
	mux.HandleFunc("/traceroute", tracerouteHandler)
	mux.HandleFunc("/traceroute6", tracerouteHandler)

	for _, listenAddr := range setting.listen {
		go func(addr string) {
			fmt.Printf("Listening on %s...\n", addr)

			var l net.Listener
			var err error

			if strings.HasPrefix(addr, "/") {
				// Delete existing socket file, ignore errors (will fail later anyway)
				os.Remove(addr)
				l, err = net.Listen("unix", addr)
			} else {
				if !strings.Contains(addr, ":") {
					addr = ":" + addr
				}
				l, err = net.Listen("tcp", addr)
			}

			if err != nil {
				panic(err)
			}

			http.Serve(l, handlers.LoggingHandler(os.Stdout, accessHandler(mux)))
		}(listenAddr)
	}

	select {}
}
