package main

import (
    "os"
    "flag"
    "strings"
    "strconv"
)

var settingServers []string
var settingServersDomain string
var settingServersPort int
var settingWhoisServer string
var settingListen string

func main() {
    serversDefault := ""
    domainDefault := ""
    proxyPortDefault := 8000
    whoisDefault := "whois.verisign-grs.com"
    listenDefault := ":5000"

    if serversEnv := os.Getenv("BIRDLG_SERVERS"); serversEnv != "" {
        serversDefault = serversEnv
    }
    if domainEnv := os.Getenv("BIRDLG_DOMAIN"); domainEnv != "" {
        domainDefault = domainEnv
    }
    if proxyPortEnv := os.Getenv("BIRDLG_PROXY_PORT"); proxyPortEnv != "" {
        var err error
        proxyPortDefault, err = strconv.Atoi(proxyPortEnv)
        if err != nil{
            panic(err)
        }
    }
    if whoisEnv := os.Getenv("BIRDLG_WHOIS"); whoisEnv != "" {
        whoisDefault = whoisEnv
    }
    if listenEnv := os.Getenv("BIRDLG_LISTEN"); listenEnv != "" {
        listenDefault = listenEnv
    }

    serversPtr := flag.String("servers", serversDefault, "server name prefixes, separated by comma")
    domainPtr := flag.String("domain", domainDefault, "server name domain suffixes")
    proxyPortPtr := flag.Int("proxy-port", proxyPortDefault, "port bird-lgproxy is running on")
    whoisPtr := flag.String("whois", whoisDefault, "whois server for queries")
    listenPtr := flag.String("listen", listenDefault, "address bird-lg is listening on")
    flag.Parse()

    if *serversPtr == "" {
        panic("no server set")
    } else if *domainPtr == "" {
        panic("no base domain set")
    }

    settingServers = strings.Split(*serversPtr, ",")
    settingServersDomain = *domainPtr
    settingServersPort = *proxyPortPtr
    settingWhoisServer = *whoisPtr
    settingListen = *listenPtr

    webServerStart()
}
