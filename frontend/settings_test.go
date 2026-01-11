package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// resetFlags resets pflag and viper state for testing
func resetFlags() {
	// Reset pflag
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	// Reset viper
	viper.Reset()
}

// TestFlagNormalizer tests that the flag normalizer converts underscores to dashes
func TestFlagNormalizer(t *testing.T) {
	resetFlags()

	// Set up the normalizer
	pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	})

	// Define a test flag
	pflag.String("test-flag", "default", "test flag")

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "dash format",
			args:     []string{"--test-flag=dash-value"},
			expected: "dash-value",
		},
		{
			name:     "underscore format",
			args:     []string{"--test_flag=underscore-value"},
			expected: "underscore-value",
		},
		{
			name:     "mixed format",
			args:     []string{"--test_flag=mixed-value"},
			expected: "mixed-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag value
			resetFlags()
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})
			pflag.String("test-flag", "default", "test flag")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Get the value
			val, err := pflag.CommandLine.GetString("test-flag")
			if err != nil {
				t.Fatalf("Failed to get flag value: %v", err)
			}

			if val != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, val)
			}
		})
	}
}

// TestServerConfigNormalization tests that server configuration flags accept both formats
func TestServerConfigNormalization(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedServers string
		expectedDomain  string
		expectedProxy   int
		expectedWhois   string
	}{
		{
			name: "dash format",
			args: []string{
				"--servers=server1,server2",
				"--domain=example.com",
				"--proxy-port=9000",
				"--whois=whois.example.com",
			},
			expectedServers: "server1,server2",
			expectedDomain:  "example.com",
			expectedProxy:   9000,
			expectedWhois:   "whois.example.com",
		},
		{
			name: "underscore format",
			args: []string{
				"--servers=srv1,srv2",
				"--domain=test.org",
				"--proxy_port=8080",
				"--whois=whois.test.org",
			},
			expectedServers: "srv1,srv2",
			expectedDomain:  "test.org",
			expectedProxy:   8080,
			expectedWhois:   "whois.test.org",
		},
		{
			name: "mixed format",
			args: []string{
				"--servers=node1,node2",
				"--domain=mixed.net",
				"--proxy-port=7000",
				"--whois=whois.mixed.net",
			},
			expectedServers: "node1,node2",
			expectedDomain:  "mixed.net",
			expectedProxy:   7000,
			expectedWhois:   "whois.mixed.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Set up the normalizer
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})

			// Define flags (using dash format as the standard)
			pflag.String("servers", "", "server names")
			pflag.String("domain", "", "domain suffix")
			pflag.Int("proxy-port", 8000, "proxy port")
			pflag.String("whois", "whois.verisign-grs.com", "whois server")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Verify values
			servers, _ := pflag.CommandLine.GetString("servers")
			if servers != tt.expectedServers {
				t.Errorf("servers: expected %q, got %q", tt.expectedServers, servers)
			}

			domain, _ := pflag.CommandLine.GetString("domain")
			if domain != tt.expectedDomain {
				t.Errorf("domain: expected %q, got %q", tt.expectedDomain, domain)
			}

			proxyPort, _ := pflag.CommandLine.GetInt("proxy-port")
			if proxyPort != tt.expectedProxy {
				t.Errorf("proxy-port: expected %d, got %d", tt.expectedProxy, proxyPort)
			}

			whois, _ := pflag.CommandLine.GetString("whois")
			if whois != tt.expectedWhois {
				t.Errorf("whois: expected %q, got %q", tt.expectedWhois, whois)
			}
		})
	}
}

// TestUIConfigNormalization tests that UI configuration flags accept both formats
func TestUIConfigNormalization(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedTitle     string
		expectedNavBrand  string
		expectedNavURL    string
		expectedAllServer string
		expectedAllURL    string
	}{
		{
			name: "dash format",
			args: []string{
				"--title-brand=My LG",
				"--navbar-brand=My Network",
				"--navbar-brand-url=/home",
				"--navbar-all-servers=All Nodes",
				"--navbar-all-url=overview",
			},
			expectedTitle:     "My LG",
			expectedNavBrand:  "My Network",
			expectedNavURL:    "/home",
			expectedAllServer: "All Nodes",
			expectedAllURL:    "overview",
		},
		{
			name: "underscore format",
			args: []string{
				"--title_brand=Test LG",
				"--navbar_brand=Test Network",
				"--navbar_brand_url=/index",
				"--navbar_all_servers=All Systems",
				"--navbar_all_url=all",
			},
			expectedTitle:     "Test LG",
			expectedNavBrand:  "Test Network",
			expectedNavURL:    "/index",
			expectedAllServer: "All Systems",
			expectedAllURL:    "all",
		},
		{
			name: "mixed format",
			args: []string{
				"--title-brand=Mixed LG",
				"--navbar_brand=Mixed Net",
				"--navbar-brand-url=/main",
				"--navbar_all_servers=Everything",
				"--navbar-all-url=global",
			},
			expectedTitle:     "Mixed LG",
			expectedNavBrand:  "Mixed Net",
			expectedNavURL:    "/main",
			expectedAllServer: "Everything",
			expectedAllURL:    "global",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Set up the normalizer
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})

			// Define flags
			pflag.String("title-brand", "Bird-lg Go", "title brand")
			pflag.String("navbar-brand", "", "navbar brand")
			pflag.String("navbar-brand-url", "/", "navbar brand url")
			pflag.String("navbar-all-servers", "All Servers", "all servers text")
			pflag.String("navbar-all-url", "all", "all servers url")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Verify values
			titleBrand, _ := pflag.CommandLine.GetString("title-brand")
			if titleBrand != tt.expectedTitle {
				t.Errorf("title-brand: expected %q, got %q", tt.expectedTitle, titleBrand)
			}

			navbarBrand, _ := pflag.CommandLine.GetString("navbar-brand")
			if navbarBrand != tt.expectedNavBrand {
				t.Errorf("navbar-brand: expected %q, got %q", tt.expectedNavBrand, navbarBrand)
			}

			navbarBrandURL, _ := pflag.CommandLine.GetString("navbar-brand-url")
			if navbarBrandURL != tt.expectedNavURL {
				t.Errorf("navbar-brand-url: expected %q, got %q", tt.expectedNavURL, navbarBrandURL)
			}

			navbarAllServers, _ := pflag.CommandLine.GetString("navbar-all-servers")
			if navbarAllServers != tt.expectedAllServer {
				t.Errorf("navbar-all-servers: expected %q, got %q", tt.expectedAllServer, navbarAllServers)
			}

			navbarAllURL, _ := pflag.CommandLine.GetString("navbar-all-url")
			if navbarAllURL != tt.expectedAllURL {
				t.Errorf("navbar-all-url: expected %q, got %q", tt.expectedAllURL, navbarAllURL)
			}
		})
	}
}

// TestNetworkConfigNormalization tests that network configuration flags accept both formats
func TestNetworkConfigNormalization(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedDNS     string
		expectedNetMode string
		expectedBgpmap  string
	}{
		{
			name: "dash format",
			args: []string{
				"--dns-interface=asn.cymru.com",
				"--net-specific-mode=dn42",
				"--bgpmap-info=asn,as-name",
			},
			expectedDNS:     "asn.cymru.com",
			expectedNetMode: "dn42",
			expectedBgpmap:  "asn,as-name",
		},
		{
			name: "underscore format",
			args: []string{
				"--dns_interface=origin.asn.cymru.com",
				"--net_specific_mode=",
				"--bgpmap_info=asn,descr",
			},
			expectedDNS:     "origin.asn.cymru.com",
			expectedNetMode: "",
			expectedBgpmap:  "asn,descr",
		},
		{
			name: "mixed format",
			args: []string{
				"--dns-interface=test.cymru.com",
				"--net_specific_mode=dn42",
				"--bgpmap-info=asn,ASName",
			},
			expectedDNS:     "test.cymru.com",
			expectedNetMode: "dn42",
			expectedBgpmap:  "asn,ASName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Set up the normalizer
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})

			// Define flags
			pflag.String("dns-interface", "asn.cymru.com", "dns interface")
			pflag.String("net-specific-mode", "", "network mode")
			pflag.String("bgpmap-info", "asn,as-name,ASName,descr", "bgpmap info")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Verify values
			dnsInterface, _ := pflag.CommandLine.GetString("dns-interface")
			if dnsInterface != tt.expectedDNS {
				t.Errorf("dns-interface: expected %q, got %q", tt.expectedDNS, dnsInterface)
			}

			netMode, _ := pflag.CommandLine.GetString("net-specific-mode")
			if netMode != tt.expectedNetMode {
				t.Errorf("net-specific-mode: expected %q, got %q", tt.expectedNetMode, netMode)
			}

			bgpmapInfo, _ := pflag.CommandLine.GetString("bgpmap-info")
			if bgpmapInfo != tt.expectedBgpmap {
				t.Errorf("bgpmap-info: expected %q, got %q", tt.expectedBgpmap, bgpmapInfo)
			}
		})
	}
}

// TestFilterConfigNormalization tests that filter configuration flags accept both formats
func TestFilterConfigNormalization(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectedProtocol    string
		expectedName        string
		expectedTelegramBot string
	}{
		{
			name: "dash format",
			args: []string{
				"--protocol-filter=bgp,ospf",
				"--name-filter=test.*",
				"--telegram-bot-name=mybot",
			},
			expectedProtocol:    "bgp,ospf",
			expectedName:        "test.*",
			expectedTelegramBot: "mybot",
		},
		{
			name: "underscore format",
			args: []string{
				"--protocol_filter=static,kernel",
				"--name_filter=ignore_.*",
				"--telegram_bot_name=testbot",
			},
			expectedProtocol:    "static,kernel",
			expectedName:        "ignore_.*",
			expectedTelegramBot: "testbot",
		},
		{
			name: "mixed format",
			args: []string{
				"--protocol-filter=bgp",
				"--name_filter=^temp",
				"--telegram-bot-name=prodbot",
			},
			expectedProtocol:    "bgp",
			expectedName:        "^temp",
			expectedTelegramBot: "prodbot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Set up the normalizer
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})

			// Define flags
			pflag.String("protocol-filter", "", "protocol filter")
			pflag.String("name-filter", "", "name filter")
			pflag.String("telegram-bot-name", "", "telegram bot name")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Verify values
			protocolFilter, _ := pflag.CommandLine.GetString("protocol-filter")
			if protocolFilter != tt.expectedProtocol {
				t.Errorf("protocol-filter: expected %q, got %q", tt.expectedProtocol, protocolFilter)
			}

			nameFilter, _ := pflag.CommandLine.GetString("name-filter")
			if nameFilter != tt.expectedName {
				t.Errorf("name-filter: expected %q, got %q", tt.expectedName, nameFilter)
			}

			telegramBotName, _ := pflag.CommandLine.GetString("telegram-bot-name")
			if telegramBotName != tt.expectedTelegramBot {
				t.Errorf("telegram-bot-name: expected %q, got %q", tt.expectedTelegramBot, telegramBotName)
			}
		})
	}
}

// TestTimeoutConfigNormalization tests that timeout configuration flags accept both formats
func TestTimeoutConfigNormalization(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectedTimeout     int
		expectedConnTimeout int
		expectedTrustProxy  bool
	}{
		{
			name: "dash format",
			args: []string{
				"--time-out=60",
				"--connection-time-out=10",
				"--trust-proxy-headers=true",
			},
			expectedTimeout:     60,
			expectedConnTimeout: 10,
			expectedTrustProxy:  true,
		},
		{
			name: "underscore format",
			args: []string{
				"--time_out=180",
				"--connection_time_out=15",
				"--trust_proxy_headers=false",
			},
			expectedTimeout:     180,
			expectedConnTimeout: 15,
			expectedTrustProxy:  false,
		},
		{
			name: "mixed format",
			args: []string{
				"--time-out=90",
				"--connection_time_out=8",
				"--trust-proxy-headers=true",
			},
			expectedTimeout:     90,
			expectedConnTimeout: 8,
			expectedTrustProxy:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Set up the normalizer
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})

			// Define flags
			pflag.Int("time-out", 120, "timeout")
			pflag.Int("connection-time-out", 5, "connection timeout")
			pflag.Bool("trust-proxy-headers", false, "trust proxy headers")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Verify values
			timeout, _ := pflag.CommandLine.GetInt("time-out")
			if timeout != tt.expectedTimeout {
				t.Errorf("time-out: expected %d, got %d", tt.expectedTimeout, timeout)
			}

			connTimeout, _ := pflag.CommandLine.GetInt("connection-time-out")
			if connTimeout != tt.expectedConnTimeout {
				t.Errorf("connection-time-out: expected %d, got %d", tt.expectedConnTimeout, connTimeout)
			}

			trustProxy, _ := pflag.CommandLine.GetBool("trust-proxy-headers")
			if trustProxy != tt.expectedTrustProxy {
				t.Errorf("trust-proxy-headers: expected %v, got %v", tt.expectedTrustProxy, trustProxy)
			}
		})
	}
}

func TestParseSettings(t *testing.T) {
	resetFlags()
	parseSettings()
	resetFlags()
}
