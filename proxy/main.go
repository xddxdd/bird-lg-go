package main

import (
	"flag"
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
}

var setting settingType

// Wrapper of tracer
func main() {
	// Prepare default socket paths, use environment variable if possible
	var settingDefault = settingType{
		"/var/run/bird/bird.ctl",
		"8000",
		[]string{""},
		"traceroute",
	}

	if birdSocketEnv := os.Getenv("BIRD_SOCKET"); birdSocketEnv != "" {
		settingDefault.birdSocket = birdSocketEnv
	}
	if listenEnv := os.Getenv("BIRDLG_LISTEN"); listenEnv != "" {
		settingDefault.listen = listenEnv
	}
	if listenEnv := os.Getenv("BIRDLG_PROXY_PORT"); listenEnv != "" {
		settingDefault.listen = listenEnv
	}
	if AllowedIPsEnv := os.Getenv("ALLOWED_IPS"); AllowedIPsEnv != "" {
		settingDefault.allowedIPs = strings.Split(AllowedIPsEnv, ",")
	}
	if tr_binEnv := os.Getenv("BIRDLG_TRACEROUTE_BIN"); tr_binEnv != "" {
		settingDefault.tr_bin = tr_binEnv
	}

	// Allow parameters to override environment variables
	birdParam := flag.String("bird", settingDefault.birdSocket, "socket file for bird, set either in parameter or environment variable BIRD_SOCKET")
	listenParam := flag.String("listen", settingDefault.listen, "listen address, set either in parameter or environment variable BIRDLG_PROXY_PORT")
	AllowedIPsParam := flag.String("allowed", strings.Join(settingDefault.allowedIPs, ","), "IPs allowed to access this proxy, separated by commas. Don't set to allow all IPs.")
	tr_binParam := flag.String("traceroute_bin", settingDefault.tr_bin, "traceroute binary file, set either in parameter or environment variable BIRDLG_TRACEROUTE_BIN")
	flag.Parse()

	if !strings.Contains(*listenParam, ":") {
		listenHost := ":" + (*listenParam)
		listenParam = &listenHost
	}

	setting.birdSocket = *birdParam
	setting.listen = *listenParam
	setting.allowedIPs = strings.Split(*AllowedIPsParam, ",")
	setting.tr_bin = *tr_binParam

	// Start HTTP server
	http.HandleFunc("/", invalidHandler)
	http.HandleFunc("/bird", birdHandler)
	http.HandleFunc("/bird6", birdHandler)
	http.HandleFunc("/traceroute", tracerouteHandler)
	http.HandleFunc("/traceroute6", tracerouteHandler)
	http.ListenAndServe(*listenParam, handlers.LoggingHandler(os.Stdout, accessHandler(http.DefaultServeMux)))
}
