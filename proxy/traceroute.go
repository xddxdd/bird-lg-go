package main

import (
	"net/http"
	"os/exec"
	"runtime"
)

// Wrapper of traceroute, IPv4
func tracerouteIPv4Wrapper(httpW http.ResponseWriter, httpR *http.Request) {
	tracerouteRealHandler(false, httpW, httpR)
}

// Wrapper of traceroute, IPv6
func tracerouteIPv6Wrapper(httpW http.ResponseWriter, httpR *http.Request) {
	tracerouteRealHandler(true, httpW, httpR)
}

// Real handler of traceroute requests
func tracerouteRealHandler(useIPv6 bool, httpW http.ResponseWriter, httpR *http.Request) {
	query := string(httpR.URL.Query().Get("q"))
	if query == "" {
		invalidHandler(httpW, httpR)
	} else {
		var cmd string
		var args []string
		if runtime.GOOS == "freebsd" || runtime.GOOS == "netbsd" {
			if useIPv6 {
				cmd = "traceroute6"
			} else {
				cmd = "traceroute"
			}
			args = []string{"-a", "-q1", "-w1", "-m15", query}
		} else if runtime.GOOS == "openbsd" {
			if useIPv6 {
				cmd = "traceroute6"
			} else {
				cmd = "traceroute"
			}
			args = []string{"-A", "-q1", "-w1", "-m15", query}
		} else if runtime.GOOS == "linux" {
			cmd = "traceroute"
			if useIPv6 {
				args = []string{"-6", "-A", "-q1", "-N32", "-w1", "-m15", query}
			} else {
				args = []string{"-4", "-A", "-q1", "-N32", "-w1", "-m15", query}
			}
		} else {
			httpW.WriteHeader(http.StatusInternalServerError)
			httpW.Write([]byte("Traceroute Not Supported\n"))
			return
		}
		instance := exec.Command(cmd, args...)
		output, err := instance.Output()
		if err != nil && runtime.GOOS == "linux" {
			// Standard traceroute utility failed, maybe system using busybox
			// Run with less parameters
			cmd = "traceroute"
			if useIPv6 {
				args = []string{"-6", "-q1", "-w1", "-m15", query}
			} else {
				args = []string{"-4", "-q1", "-w1", "-m15", query}
			}
			instance = exec.Command(cmd, args...)
			output, err = instance.Output()
		}
		if err != nil {
			httpW.WriteHeader(http.StatusInternalServerError)
			httpW.Write([]byte("Traceroute Execution Error: "))
			httpW.Write([]byte(err.Error() + "\n"))
			return
		}
		httpW.Write(output)
	}
}
