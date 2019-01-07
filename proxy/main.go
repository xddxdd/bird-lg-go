package main

import (
    "io"
    "net"
    "net/http"
    "sync"
    "runtime"
    "os/exec"
    "flag"
)

// BIRDv4 connection & mutex lock
var bird net.Conn
var birdMutex = &sync.Mutex{}

// BIRDv6 connection & mutex lock
var bird6 net.Conn
var bird6Mutex = &sync.Mutex{}

// Check if a byte is character for number
func isNumeric(b byte) bool {
    return b >= byte('0') && b <= byte('9')
}

// Read a line from bird socket, removing preceding status number, output it.
// Returns if there are more lines.
func birdReadln(bird io.Reader, w io.Writer) bool {
    // Read from socket byte by byte, until reaching newline character
    c := make([]byte, 1024, 1024)
    pos := 0
    for {
        if pos >= 1024 { break }
        _, err := bird.Read(c[pos:pos+1])
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
            for c[pos] == byte(' ') { pos++ }
            w.Write(c[pos:])
        }
        return true
    } else if isNumeric(c[0]) && isNumeric(c[1]) && isNumeric(c[2]) && isNumeric(c[3]) {
        // There is a status number at beginning, remove first 5 bytes
        if w != nil && pos > 6 {
            pos = 5
            for c[pos] == byte(' ') { pos++ }
            w.Write(c[pos:])
        }
        return c[0] != byte('0') && c[0] != byte('8') && c[0] != byte('9')
    } else {
        // There is no status number, only remove preceding spaces
        if w != nil {
            pos = 0
            for c[pos] == byte(' ') { pos++ }
            w.Write(c[pos:])
        }
        return true
    }
}

// Write a command to a bird socket
func birdWriteln(bird io.Writer, s string) {
    bird.Write([]byte(s + "\n"))
}

// Default handler, returns 500 Internal Server Error
func invalidHandler(httpW http.ResponseWriter, httpR *http.Request) {
    httpW.WriteHeader(http.StatusInternalServerError)
    httpW.Write([]byte("Invalid Request\n"))
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
        for birdReadln(bird, httpW) {}
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
        for birdReadln(bird6, httpW) {}
    }
}

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
            if useIPv6 { cmd = "traceroute6" } else { cmd = "traceroute" }
            args = []string{"-a", "-q1", "-w1", "-m15", query}
        } else if runtime.GOOS == "openbsd" {
            if useIPv6 { cmd = "traceroute6" } else { cmd = "traceroute" }
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
        if err != nil {
            httpW.WriteHeader(http.StatusInternalServerError)
            httpW.Write([]byte("Traceroute Execution Error: "))
            httpW.Write([]byte(err.Error() + "\n"))
            return
        }
        httpW.Write(output)
    }
}

func main() {
    var err error

    birdPtr := flag.String("bird", "/var/run/bird/bird.ctl", "socket file for bird")
    bird6Ptr := flag.String("bird6", "/var/run/bird/bird6.ctl", "socket file for bird6")
    flag.Parse()

    // Initialize BIRDv4 socket
    bird, err = net.Dial("unix", *birdPtr)
    if err != nil {
        panic(err)
    }
    defer bird.Close()

    birdReadln(bird, nil)
    birdWriteln(bird, "restrict")
    birdReadln(bird, nil)

    // Initialize BIRDv6 socket
    bird6, err = net.Dial("unix", *bird6Ptr)
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
    http.ListenAndServe(":8000", nil)
}
