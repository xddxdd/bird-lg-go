package main

import (
    "net"
    "io/ioutil"
)

// Send a whois request
func whois(s string) string {
	conn, err := net.Dial("tcp", settingWhoisServer + ":43")
	if err != nil {
		return err.Error()
	}
	defer conn.Close()

    conn.Write([]byte(s + "\r\n"))
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		return err.Error()
	}
    return string(result)
}
