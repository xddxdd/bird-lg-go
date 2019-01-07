package main

import (
    "flag"
    "strings"
)

var settingServers []string
var settingServersDomain string
var settingServersPort int
var settingWhoisServer string

func main() {
    serversPtr := flag.String("servers", "", "server name prefixes, separated by comma")
    domainPtr := flag.String("domain", "", "server name domain suffixes")
    portPtr := flag.Int("port", 8000, "port bird-lgproxy is running on")
    whoisPtr := flag.String("whois", "whois.verisign-grs.com", "whois server for queries")
    flag.Parse()

    settingServers = strings.Split(*serversPtr, ",")
    settingServersDomain = *domainPtr
    settingServersPort = *portPtr
    settingWhoisServer = *whoisPtr
    webServerStart()
}
