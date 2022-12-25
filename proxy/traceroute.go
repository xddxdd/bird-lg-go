package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

func tracerouteArgsToString(cmd string, args []string, target []string) string {
	var cmdCombined = append([]string{cmd}, args...)
	cmdCombined = append(cmdCombined, target...)
	return strings.Join(cmdCombined, " ")
}

func tracerouteTryExecute(cmd string, args []string, target []string) ([]byte, error) {
	instance := exec.Command(cmd, append(args, target...)...)
	output, err := instance.CombinedOutput()
	if err == nil {
		return output, nil
	}

	return output, err
}

func tracerouteDetect(cmd string, args []string) bool {
	target := []string{"127.0.0.1"}
	success := false
	if result, err := tracerouteTryExecute(cmd, args, target); err == nil {
		setting.tr_bin = cmd
		setting.tr_flags = args
		success = true
		fmt.Printf("Traceroute autodetect success: %s\n", tracerouteArgsToString(cmd, args, target))
	} else {
		fmt.Printf("Traceroute autodetect fail, continuing: %s (%s)\n%s", tracerouteArgsToString(cmd, args, target), err.Error(), result)
	}

	return success
}

func tracerouteAutodetect() {
	if setting.tr_bin != "" && setting.tr_flags != nil {
		return
	}

	// Traceroute (custom binary)
	if setting.tr_bin != "" {
		if tracerouteDetect(setting.tr_bin, []string{"-q1", "-N32", "-w1"}) {
			return
		}
		if tracerouteDetect(setting.tr_bin, []string{"-q1", "-w1"}) {
			return
		}
		if tracerouteDetect(setting.tr_bin, []string{}) {
			return
		}
	}

	// MTR
	if tracerouteDetect("mtr", []string{"-w", "-c1", "-Z1", "-G1", "-b"}) {
		return
	}

	// Traceroute
	if tracerouteDetect("traceroute", []string{"-q1", "-N32", "-w1"}) {
		return
	}
	if tracerouteDetect("traceroute", []string{"-q1", "-w1"}) {
		return
	}
	if tracerouteDetect("traceroute", []string{}) {
		return
	}

	// Unsupported
	setting.tr_bin = ""
	setting.tr_flags = nil
	println("Traceroute autodetect failed! Traceroute will be disabled")
}

func tracerouteHandler(httpW http.ResponseWriter, httpR *http.Request) {
	query := string(httpR.URL.Query().Get("q"))
	query = strings.TrimSpace(query)

	if query == "" {
		invalidHandler(httpW, httpR)
	} else {
		args, err := shlex.Split(query)
		if err != nil {
			httpW.WriteHeader(http.StatusInternalServerError)
			httpW.Write([]byte(fmt.Sprintf("failed to parse args: %s\n", err.Error())))
			return
		}

		var result []byte
		skippedCounter := 0

		if setting.tr_bin == "" {
			httpW.WriteHeader(http.StatusInternalServerError)
			httpW.Write([]byte("traceroute not supported on this node.\n"))
			return
		}

		result, err = tracerouteTryExecute(setting.tr_bin, setting.tr_flags, args)
		if err != nil {
			httpW.WriteHeader(http.StatusInternalServerError)
			httpW.Write([]byte(fmt.Sprintf("Error executing traceroute: %s\n\n", err.Error())))
		}

		if result != nil {
			if setting.tr_raw {
				httpW.Write(result)
			} else {
				resultString := string(result)
				resultString = regexp.MustCompile(`(?m)^\s*(\d*)\s*\*\n`).ReplaceAllStringFunc(resultString, func(w string) string {
					skippedCounter++
					return ""
				})
				httpW.Write([]byte(strings.TrimSpace(resultString)))
				if skippedCounter > 0 {
					httpW.Write([]byte("\n\n" + strconv.Itoa(skippedCounter) + " hops not responding."))
				}
			}
		}
	}
}
