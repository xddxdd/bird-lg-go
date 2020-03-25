package main

import (
	"io"
	"net"
	"net/http"
	"sync"
)

// BIRDv4 connection & mutex lock
var bird net.Conn
var birdMutex = &sync.Mutex{}

// BIRDv6 connection & mutex lock
var bird6 net.Conn
var bird6Mutex = &sync.Mutex{}

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
	if pos < 4 {
		// Line is too short to have a status number
		if w != nil {
			pos = 0
			for c[pos] == byte(' ') {
				pos++
			}
			w.Write(c[pos:])
		}
		return true
	} else if isNumeric(c[0]) && isNumeric(c[1]) && isNumeric(c[2]) && isNumeric(c[3]) {
		// There is a status number at beginning, remove first 5 bytes
		if w != nil && pos > 6 {
			pos = 5
			for c[pos] == byte(' ') {
				pos++
			}
			w.Write(c[pos:])
		}
		return c[0] != byte('0') && c[0] != byte('8') && c[0] != byte('9')
	} else {
		// There is no status number, only remove preceding spaces
		if w != nil {
			pos = 0
			for c[pos] == byte(' ') {
				pos++
			}
			w.Write(c[pos:])
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
		birdMutex.Lock()
		defer birdMutex.Unlock()

		println(query)
		birdWriteln(bird, query)
		for birdReadln(bird, httpW) {
		}
	}
}

// Handles BIRDv6 queries
func bird6Handler(httpW http.ResponseWriter, httpR *http.Request) {
	query := string(httpR.URL.Query().Get("q"))
	if query == "" {
		invalidHandler(httpW, httpR)
	} else {
		bird6Mutex.Lock()
		defer bird6Mutex.Unlock()

		println(query)
		birdWriteln(bird6, query)
		for birdReadln(bird6, httpW) {
		}
	}
}
