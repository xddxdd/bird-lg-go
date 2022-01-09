package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
)

// Read a line from bird socket, removing preceding status number, output it.
// Returns if there are more lines.
func birdReadln(bird io.Reader, w io.Writer) bool {
	// Read from socket byte by byte, until reaching newline character
	c := make([]byte, 1024, 1024)
	pos := 0
	for {
		if pos >= 1024 {
			break
		}
		_, err := bird.Read(c[pos : pos+1])
		if err != nil {
			panic(err)
		}
		if c[pos] == byte('\n') {
			break
		}
		pos++
	}

	c = c[:pos+1]
	// print(string(c[:]))

	// Remove preceding status number, different situations
	if pos > 4 && isNumeric(c[0]) && isNumeric(c[1]) && isNumeric(c[2]) && isNumeric(c[3]) {
		// There is a status number at beginning, remove first 5 bytes
		if w != nil && pos > 6 {
			pos = 5
			w.Write(c[pos:])
		}
		return c[0] != byte('0') && c[0] != byte('8') && c[0] != byte('9')
	} else {
		if w != nil {
			w.Write(c[1:])
		}
		return true
	}
}

// Write a command to a bird socket
func birdWriteln(bird io.Writer, s string) {
	bird.Write([]byte(s + "\n"))
}

// Handles BIRDv4 queries
func birdHandler(httpW http.ResponseWriter, httpR *http.Request) {
	query := string(httpR.URL.Query().Get("q"))
	if query == "" {
		invalidHandler(httpW, httpR)
	} else {
		// Initialize BIRDv4 socket
		bird, err := net.Dial("unix", setting.birdSocket)
		if err != nil {
			httpW.WriteHeader(http.StatusInternalServerError)
			httpW.Write([]byte(err.Error()))
			return
		}
		defer bird.Close()

		birdReadln(bird, nil)
		birdWriteln(bird, "restrict")
		var restrictedConfirmation bytes.Buffer
		birdReadln(bird, &restrictedConfirmation)
		if !strings.Contains(restrictedConfirmation.String(), "Access restricted") {
			httpW.WriteHeader(http.StatusInternalServerError)
			httpW.Write([]byte("could not verify that bird access was restricted"))
			return
		}
		birdWriteln(bird, query)
		for birdReadln(bird, httpW) {
		}
	}
}
