package main

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jarcoal/httpmock"
)

type channelData struct {
	id   int
	data string
}

func createConnectionTimeoutRoundTripper(timeout int) http.RoundTripper {
	context := net.Dialer{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Prefer httpmock's transport if activated, so unit tests can work
	if http.DefaultTransport == httpmock.DefaultTransport {
		return httpmock.DefaultTransport
	}

	return &http.Transport{
		DialContext: context.DialContext,

		// Default options from transport.go
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// Send commands to lgproxy instances in parallel, and retrieve their responses
func batchRequest(servers []string, endpoint string, command string) []string {
	if len(servers) > len(setting.servers) {
		return []string{"invalid request: too many servers specified"}
	}

	// Channel and array for storing responses
	var ch chan channelData = make(chan channelData)
	var responseArray []string = make([]string, len(servers))
	var createdGoroutines int = 0

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
			// If the server is not valid, return a failure
			responseArray[i] = "request failed: invalid server\n"
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
				client := http.Client{
					Transport: createConnectionTimeoutRoundTripper(setting.connectionTimeOut),
					Timeout:   time.Duration(setting.timeOut) * time.Second,
				}
				response, err := client.Get(url)
				if err != nil {
					ch <- channelData{i, "request failed: " + err.Error() + "\n"}
					return
				}

				buf := make([]byte, 65536)
				n, err := io.ReadFull(response.Body, buf)
				if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
					ch <- channelData{i, "request failed: " + err.Error()}
				} else {
					ch <- channelData{i, string(buf[:n])}
				}
			}(url, i)
			createdGoroutines++
		}
	}

	// Sort the responses by their ids, to return data in order
	for i := 0; i < createdGoroutines; i++ {
		var output channelData = <-ch
		responseArray[output.id] = output.data
		if len(responseArray[output.id]) == 0 {
			responseArray[output.id] = "node returned empty response, please refresh to try again."
		}
	}

	return responseArray
}
