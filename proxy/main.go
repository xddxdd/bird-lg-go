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
	birdSocket  string
	bird6Socket string
	listen      string
	allowedIPs  []string
}

var setting settingType

// Wrapper of tracer
func main() {
	// Prepare default socket paths, use environment variable if possible
	var settingDefault = settingType{
		"/var/run/bird/bird.ctl",
		"/var/run/bird/bird6.ctl",
		":8000",
		[]string{""},
	}

	if birdSocketEnv := os.Getenv("BIRD_SOCKET"); birdSocketEnv != "" {
		settingDefault.birdSocket = birdSocketEnv
	}
	if bird6SocketEnv := os.Getenv("BIRD6_SOCKET"); bird6SocketEnv != "" {
		settingDefault.bird6Socket = bird6SocketEnv
	}
	if listenEnv := os.Getenv("BIRDLG_LISTEN"); listenEnv != "" {
		settingDefault.listen = listenEnv
	}
	if AllowedIPsEnv := os.Getenv("ALLOWED_IPS"); AllowedIPsEnv != "" {
		settingDefault.allowedIPs = strings.Split(AllowedIPsEnv, ",")
	}

	// Allow parameters to override environment variables
	birdParam := flag.String("bird", settingDefault.birdSocket, "socket file for bird, set either in parameter or environment variable BIRD_SOCKET")
	bird6Param := flag.String("bird6", settingDefault.bird6Socket, "socket file for bird6, set either in parameter or environment variable BIRD6_SOCKET")
	listenParam := flag.String("listen", settingDefault.listen, "listen address, set either in parameter or environment variable BIRDLG_LISTEN")
	AllowedIPsParam := flag.String("allowed", strings.Join(settingDefault.allowedIPs, ","), "IPs allowed to access this proxy, separated by commas. Don't set to allow all IPs.")
	flag.Parse()

	setting.birdSocket = *birdParam
	setting.bird6Socket = *bird6Param
	setting.listen = *listenParam
	setting.allowedIPs = strings.Split(*AllowedIPsParam, ",")

	// Start HTTP server
	http.HandleFunc("/", invalidHandler)
	http.HandleFunc("/bird", birdHandler)
	http.HandleFunc("/bird6", bird6Handler)
	http.HandleFunc("/traceroute", tracerouteIPv4Wrapper)
	http.HandleFunc("/traceroute6", tracerouteIPv6Wrapper)
	http.ListenAndServe(*listenParam, handlers.LoggingHandler(os.Stdout, accessHandler(http.DefaultServeMux)))
}
