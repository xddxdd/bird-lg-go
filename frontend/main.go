package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type settingType struct {
	servers         []string
	serversDisplay  []string
	domain          string
	proxyPort       int
	whoisServer     string
	listen          string
	dnsInterface    string
	netSpecificMode string
	titleBrand      string
	navBarBrand     string
	navBarBrandURL  string
	navBarAllServer string
	navBarAllURL    string
	bgpmapInfo      string
	telegramBotName string
	protocolFilter  []string
}

var setting settingType

func main() {
	var settingDefault = settingType{
		servers:         []string{""},
		proxyPort:       8000,
		whoisServer:     "whois.verisign-grs.com",
		listen:          "5000",
		dnsInterface:    "asn.cymru.com",
		titleBrand:      "Bird-lg Go",
		navBarBrand:     "Bird-lg Go",
		navBarBrandURL:  "/",
		navBarAllServer: "All Servers",
		navBarAllURL:    "all",
		bgpmapInfo:      "asn,as-name,ASName,descr",
		telegramBotName: "",
		protocolFilter:  []string{},
	}

	if env := os.Getenv("BIRDLG_SERVERS"); env != "" {
		settingDefault.servers = strings.Split(env, ",")
	}
	if env := os.Getenv("BIRDLG_DOMAIN"); env != "" {
		settingDefault.domain = env
	}
	if env := os.Getenv("BIRDLG_PROXY_PORT"); env != "" {
		var err error
		if settingDefault.proxyPort, err = strconv.Atoi(env); err != nil {
			panic(err)
		}
	}
	if env := os.Getenv("BIRDLG_WHOIS"); env != "" {
		settingDefault.whoisServer = env
	}
	if env := os.Getenv("BIRDLG_LISTEN"); env != "" {
		settingDefault.listen = env
	}
	if env := os.Getenv("BIRDLG_DNS_INTERFACE"); env != "" {
		settingDefault.dnsInterface = env
	}
	if env := os.Getenv("BIRDLG_NET_SPECIFIC_MODE"); env != "" {
		settingDefault.netSpecificMode = env
	}
	if env := os.Getenv("BIRDLG_TITLE_BRAND"); env != "" {
		settingDefault.titleBrand = env
		settingDefault.navBarBrand = env
	}
	if env := os.Getenv("BIRDLG_NAVBAR_BRAND"); env != "" {
		settingDefault.navBarBrand = env
	}
	if env := os.Getenv("BIRDLG_NAVBAR_BRAND_URL"); env != "" {
		settingDefault.navBarBrandURL = env
	}
	if env := os.Getenv("BIRDLG_NAVBAR_ALL_SERVERS"); env != "" {
		settingDefault.navBarAllServer = env
	}
	if env := os.Getenv("BIRDLG_NAVBAR_ALL_URL"); env != "" {
		settingDefault.navBarAllURL = env
	}
	if env := os.Getenv("BIRDLG_BGPMAP_INFO"); env != "" {
		settingDefault.bgpmapInfo = env
	}
	if env := os.Getenv("BIRDLG_TELEGRAM_BOT_NAME"); env != "" {
		settingDefault.telegramBotName = env
	}
	if env := os.Getenv("BIRDLG_PROTOCOL_FILTER"); env != "" {
		settingDefault.protocolFilter = strings.Split(env, ",")
	}

	serversPtr := flag.String("servers", strings.Join(settingDefault.servers, ","), "server name prefixes, separated by comma")
	domainPtr := flag.String("domain", settingDefault.domain, "server name domain suffixes")
	proxyPortPtr := flag.Int("proxy-port", settingDefault.proxyPort, "port bird-lgproxy is running on")
	whoisPtr := flag.String("whois", settingDefault.whoisServer, "whois server for queries")
	listenPtr := flag.String("listen", settingDefault.listen, "address bird-lg is listening on")
	dnsInterfacePtr := flag.String("dns-interface", settingDefault.dnsInterface, "dns zone to query ASN information")
	netSpecificModePtr := flag.String("net-specific-mode", settingDefault.netSpecificMode, "network specific operation mode, [(none)|dn42]")
	titleBrandPtr := flag.String("title-brand", settingDefault.titleBrand, "prefix of page titles in browser tabs")
	navBarBrandPtr := flag.String("navbar-brand", settingDefault.navBarBrand, "brand to show in the navigation bar")
	navBarBrandURLPtr := flag.String("navbar-brand-url", settingDefault.navBarBrandURL, "the url of the brand to show in the navigation bar")
	navBarAllServerPtr := flag.String("navbar-all-servers", settingDefault.navBarAllServer, "the text of \"All servers\" button in the navigation bar")
	navBarAllURL := flag.String("navbar-all-url", settingDefault.navBarAllURL, "the URL of \"All servers\" button")
	bgpmapInfo := flag.String("bgpmap-info", settingDefault.bgpmapInfo, "the infos displayed in bgpmap, separated by comma, start with \":\" means allow multiline")
	telegramBotNamePtr := flag.String("telegram-bot-name", settingDefault.telegramBotName, "telegram bot name (used to filter @bot commands)")
	protocolFilterPtr := flag.String("protocol-filter", strings.Join(settingDefault.protocolFilter, ","),
		"protocol types to show in summary tables (comma separated list); defaults to all if not set")
	flag.Parse()

	if *serversPtr == "" {
		panic("no server set")
	}

	if !strings.Contains(*listenPtr, ":") {
		listenHost := ":" + (*listenPtr)
		listenPtr = &listenHost
	}

	servers := strings.Split(*serversPtr, ",")
	serversDisplay := strings.Split(*serversPtr, ",")

	protocolFilter := []string{}
	// strings.Split returns [""] for empty inputs; we want the list to remain empty in these cases
	if len(*protocolFilterPtr) > 0 {
		protocolFilter = strings.Split(*protocolFilterPtr, ",")
	}

	// Split server names of the form "DisplayName<Hostname>"
	for i, server := range servers {
		pos := strings.Index(server, "<")
		if pos != -1 {
			serversDisplay[i] = server[0:pos]
			servers[i] = server[pos+1 : len(server)-1]
		}
	}

	setting = settingType{
		servers,
		serversDisplay,
		*domainPtr,
		*proxyPortPtr,
		*whoisPtr,
		*listenPtr,
		*dnsInterfacePtr,
		strings.ToLower(*netSpecificModePtr),
		*titleBrandPtr,
		*navBarBrandPtr,
		*navBarBrandURLPtr,
		*navBarAllServerPtr,
		*navBarAllURL,
		*bgpmapInfo,
		*telegramBotNamePtr,
		protocolFilter,
	}

	ImportTemplates()
	webServerStart()
}
