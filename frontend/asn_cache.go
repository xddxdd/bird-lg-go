package main

import (
	"fmt"
	"net"
	"strings"
)

type ASNCache map[string]string

func (cache ASNCache) _lookup(asn string) string {
	// Try to get ASN representation using DNS
	if setting.dnsInterface != "" {
		records, err := net.LookupTXT(fmt.Sprintf("AS%s.%s", asn, setting.dnsInterface))
		if err == nil {
			result := strings.Join(records, " ")
			if resultSplit := strings.Split(result, " | "); len(resultSplit) > 1 {
				result = strings.Join(resultSplit[1:], "\n")
			}
			return fmt.Sprintf("AS%s\n%s", asn, result)
		}
	}

	// Try to get ASN representation using WHOIS
	if setting.whoisServer != "" {
		if setting.bgpmapInfo == "" {
			setting.bgpmapInfo = "asn,as-name,ASName,descr"
		}
		records := whois(fmt.Sprintf("AS%s", asn))
		if records != "" {
			recordsSplit := strings.Split(records, "\n")
			var result []string
			for _, title := range strings.Split(setting.bgpmapInfo, ",") {
				if title == "asn" {
					result = append(result, "AS"+asn)
				}
			}
			for _, title := range strings.Split(setting.bgpmapInfo, ",") {
				allow_multiline := false
				if title[0] == ':' && len(title) >= 2 {
					title = title[1:]
					allow_multiline = true
				}
				for _, line := range recordsSplit {
					if len(line) == 0 || line[0] == '%' || !strings.Contains(line, ":") {
						continue
					}
					linearr := strings.SplitN(line, ":", 2)
					line_title := linearr[0]
					content := strings.TrimSpace(linearr[1])
					if line_title != title {
						continue
					}
					result = append(result, content)
					if !allow_multiline {
						break
					}

				}
			}
			if len(result) > 0 {
				return strings.Join(result, "\n")
			}
		}
	}

	return ""
}

func (cache ASNCache) Lookup(asn string) string {
	cachedValue, cacheOk := cache[asn]
	if cacheOk {
		return cachedValue
	}

	result := cache._lookup(asn)
	if len(result) == 0 {
		result = fmt.Sprintf("AS%s", asn)
	}

	cache[asn] = result
	return result
}
