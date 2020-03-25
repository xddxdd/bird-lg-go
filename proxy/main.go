package main

import (
	"flag"
	"net"
	"net/http"
	"os"
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

// Wrapper of tracer
func main() {
	var err error

	// Prepare default socket paths, use environment variable if possible
	birdSocketDefault := "/var/run/bird/bird.ctl"
	bird6SocketDefault := "/var/run/bird/bird6.ctl"
	listenDefault := ":8000"

	if birdSocketEnv := os.Getenv("BIRD_SOCKET"); birdSocketEnv != "" {
		birdSocketDefault = birdSocketEnv
	}
	if bird6SocketEnv := os.Getenv("BIRD6_SOCKET"); bird6SocketEnv != "" {
		bird6SocketDefault = bird6SocketEnv
	}
	if listenEnv := os.Getenv("BIRDLG_LISTEN"); listenEnv != "" {
		listenDefault = listenEnv
	}

	// Allow parameters to override environment variables
	birdParam := flag.String("bird", birdSocketDefault, "socket file for bird, set either in parameter or environment variable BIRD_SOCKET")
	bird6Param := flag.String("bird6", bird6SocketDefault, "socket file for bird6, set either in parameter or environment variable BIRD6_SOCKET")
	listenParam := flag.String("listen", listenDefault, "listen address, set either in parameter or environment variable BIRDLG_LISTEN")
	flag.Parse()

	// Initialize BIRDv4 socket
	bird, err = net.Dial("unix", *birdParam)
	if err != nil {
		panic(err)
	}
	defer bird.Close()

	birdReadln(bird, nil)
	birdWriteln(bird, "restrict")
	birdReadln(bird, nil)

	// Initialize BIRDv6 socket
	bird6, err = net.Dial("unix", *bird6Param)
	if err != nil {
		panic(err)
	}
	defer bird6.Close()

	birdReadln(bird6, nil)
	birdWriteln(bird6, "restrict")
	birdReadln(bird6, nil)

	// Start HTTP server
	http.HandleFunc("/", invalidHandler)
	http.HandleFunc("/bird", birdHandler)
	http.HandleFunc("/bird6", bird6Handler)
	http.HandleFunc("/traceroute", tracerouteIPv4Wrapper)
	http.HandleFunc("/traceroute6", tracerouteIPv6Wrapper)
	http.ListenAndServe(*listenParam, nil)
}
