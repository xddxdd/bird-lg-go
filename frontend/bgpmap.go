package main

import (
	"fmt"
	"html"
	"net"
	"regexp"
	"strings"
)

func getASNRepresentation(asn string) string {
	if setting.dnsInterface != "" {
		// get ASN representation using DNS
		records, err := net.LookupTXT(fmt.Sprintf("AS%s.%s", asn, setting.dnsInterface))
		if err == nil {
			result := strings.Join(records, " ")
			if resultSplit := strings.Split(result, " | "); len(resultSplit) > 1 {
				result = strings.Join(resultSplit[1:], "\\n")
			}
			return fmt.Sprintf("AS%s\\n%s", asn, result)
		}
	}

	if setting.whoisServer != "" {
		// get ASN representation using WHOIS
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
				return fmt.Sprintf("%s", strings.Join(result, "\n"))
			}
		}
	}
	return fmt.Sprintf("AS%s", asn)
}

func birdRouteToGraphviz(servers []string, responses []string, target string) string {
	graph := make(map[string]string)
	// Helper to add an edge
	addEdge := func(src string, dest string, attr string) {
		key := "\"" + html.EscapeString(src) + "\" -> \"" + html.EscapeString(dest) + "\""
		_, present := graph[key]
		// If there are multiple edges / routes between 2 nodes, only pick the first one
		if present {
			return
		}
		graph[key] = attr
	}
	// Helper to set attribute for a point in graph
	addPoint := func(name string, attr string) {
		key := "\"" + html.EscapeString(name) + "\""
		_, present := graph[key]
		// Do not remove point's attributes if it's already present
		if present && len(attr) == 0 {
			return
		}
		graph[key] = attr
	}
	// The protocol name for each route (e.g. "ibgp_sea02") is encoded in the form:
	//    unicast [ibgp_sea02 2021-08-27 from fd86:bad:11b7:1::1] * (100/1015) [i]
	protocolNameRe := regexp.MustCompile(`\[(.*?) .*\]`)
	// Try to split the output into one chunk for each route.
	// Possible values are defined at https://gitlab.nic.cz/labs/bird/-/blob/v2.0.8/nest/rt-attr.c#L81-87
	routeSplitRe := regexp.MustCompile("(unicast|blackhole|unreachable|prohibited)")

	addPoint("Target: "+target, "[color=red,shape=diamond]")
	for serverID, server := range servers {
		response := responses[serverID]
		if len(response) == 0 {
			continue
		}
		addPoint(server, "[color=blue,shape=box]")
		routes := routeSplitRe.Split(response, -1)

		targetNodeName := "Target: " + target
		var nonBGPRoutes []string
		var nonBGPRoutePreferred bool

		for routeIndex, route := range routes {
			var routeNexthop string
			var routeASPath string
			var routePreferred bool = routeIndex > 0 && strings.Contains(route, "*")
			// Track non-BGP routes in the output by their protocol name, but draw them altogether in one line
			// so that there are no conflicts in the edge label
			var protocolName string

			for _, routeParameter := range strings.Split(route, "\n") {
				if strings.HasPrefix(routeParameter, "\tBGP.next_hop: ") {
					routeNexthop = strings.TrimPrefix(routeParameter, "\tBGP.next_hop: ")
				} else if strings.HasPrefix(routeParameter, "\tBGP.as_path: ") {
					routeASPath = strings.TrimPrefix(routeParameter, "\tBGP.as_path: ")
				} else {
					match := protocolNameRe.FindStringSubmatch(routeParameter)
					if len(match) >= 2 {
						protocolName = match[1]
					}
				}
			}
			if routePreferred {
				protocolName = protocolName + "*"
			}
			if len(routeASPath) == 0 {
				if routeIndex == 0 {
					// The first string split includes the target prefix and isn't a valid route
					continue
				}
				if routePreferred {
					nonBGPRoutePreferred = true
				}
				nonBGPRoutes = append(nonBGPRoutes, protocolName)
				continue
			}

			// Connect each node on AS path
			paths := strings.Split(strings.TrimSpace(routeASPath), " ")

			for pathIndex := range paths {
				paths[pathIndex] = strings.TrimPrefix(paths[pathIndex], "(")
				paths[pathIndex] = strings.TrimSuffix(paths[pathIndex], ")")
			}

			// First step starting from originating server
			if len(paths) > 0 {
				attrs := []string{"fontsize=12.0"}
				if routePreferred {
					attrs = append(attrs, "color=red")
				}
				if len(routeNexthop) > 0 {
					attrs = append(attrs, fmt.Sprintf("label=\"%s\\n%s\"", protocolName, routeNexthop))
				}
				formattedAttr := fmt.Sprintf("[%s]", strings.Join(attrs, ","))
				addEdge(server, getASNRepresentation(paths[0]), formattedAttr)
			}

			// Following steps, edges between AS
			for pathIndex := range paths {
				if pathIndex == 0 {
					continue
				}
				addEdge(getASNRepresentation(paths[pathIndex-1]), getASNRepresentation(paths[pathIndex]), (map[bool]string{true: "[color=red]"})[routePreferred])
			}
			// Last AS to destination
			addEdge(getASNRepresentation(paths[len(paths)-1]), targetNodeName, (map[bool]string{true: "[color=red]"})[routePreferred])
		}

		if len(nonBGPRoutes) > 0 {
			protocolsForRoute := fmt.Sprintf("label=\"%s\"", strings.Join(nonBGPRoutes, "\\n"))

			attrs := []string{protocolsForRoute, "fontsize=12.0"}

			if nonBGPRoutePreferred {
				attrs = append(attrs, "color=red")
			}
			formattedAttr := fmt.Sprintf("[%s]", strings.Join(attrs, ","))
			addEdge(server, targetNodeName, formattedAttr)
		}
	}

	// Combine all graphviz commands
	var result string
	for edge, attr := range graph {
		result += edge + " " + attr + ";\n"
	}
	return "digraph {\n" + result + "}\n"
}
