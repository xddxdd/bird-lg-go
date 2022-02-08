package main

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type channelData struct {
	id   int
	data string
}

// Send commands to lgproxy instances in parallel, and retrieve their responses
func batchRequest(servers []string, endpoint string, command string) []string {
	// Channel and array for storing responses
	var ch chan channelData = make(chan channelData)
	var responseArray []string = make([]string, len(servers))

	for i, server := range servers {
		// Check if the server is in the valid server list passed at startup
		var isValidServer bool = false
		for _, validServer := range setting.servers {
			if validServer == server {
				isValidServer = true
				break
			}
		}

		if !isValidServer {
			// If the server is not valid, create a dummy goroutine to return a failure
			go func(i int) {
				ch <- channelData{i, "request failed: invalid server\n"}
			}(i)
		} else {
			// Compose URL and send the request
			hostname := server
			hostname = url.PathEscape(hostname)
			if strings.Contains(hostname, ":") {
				hostname = "[" + hostname + "]"
			}
			if setting.domain != "" {
				hostname += "." + setting.domain
			}
			url := "http://" + hostname + ":" + strconv.Itoa(setting.proxyPort) + "/" + url.PathEscape(endpoint) + "?q=" + url.QueryEscape(command)
			go func(url string, i int) {
				client := http.Client{Timeout: time.Duration(setting.timeOut) * time.Second}
				response, err := client.Get(url)
				if err != nil {
					ch <- channelData{i, "request failed: " + err.Error() + "\n"}
					return
				}

				buf := make([]byte, 65536)
				n, err := io.ReadFull(response.Body, buf)
				if err != nil && err != io.ErrUnexpectedEOF {
					ch <- channelData{i, err.Error()}
				} else {
					ch <- channelData{i, string(buf[:n])}
				}
			}(url, i)
		}
	}

	// Sort the responses by their ids, to return data in order
	for range servers {
		var output channelData = <-ch
		responseArray[output.id] = output.data
		if len(responseArray[output.id]) == 0 {
			responseArray[output.id] = "node returned empty response, please refresh to try again."
		}
	}

	return responseArray
}
