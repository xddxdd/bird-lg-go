package main

import (
	"io/ioutil"
	"net"
	"time"
)

// Send a whois request
func whois(s string) string {
	if setting.whoisServer == "" {
		return ""
	}

	conn, err := net.DialTimeout("tcp", setting.whoisServer+":43", 5*time.Second)
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
