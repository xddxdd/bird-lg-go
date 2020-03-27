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

type settingType struct {
	birdSocket  string
	bird6Socket string
	listen      string
}

var setting settingType

// Wrapper of tracer
func main() {
	var err error

	// Prepare default socket paths, use environment variable if possible
	var settingDefault = settingType{
		"/var/run/bird/bird.ctl",
		"/var/run/bird/bird6.ctl",
		":8000",
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

	// Allow parameters to override environment variables
	birdParam := flag.String("bird", settingDefault.birdSocket, "socket file for bird, set either in parameter or environment variable BIRD_SOCKET")
	bird6Param := flag.String("bird6", settingDefault.bird6Socket, "socket file for bird6, set either in parameter or environment variable BIRD6_SOCKET")
	listenParam := flag.String("listen", settingDefault.listen, "listen address, set either in parameter or environment variable BIRDLG_LISTEN")
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
