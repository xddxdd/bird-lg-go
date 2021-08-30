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
		records := whois(fmt.Sprintf("AS%s", asn))
		if records != "" {
			recordsSplit := strings.Split(records, "\n")
			result := ""
			for _, line := range recordsSplit {
				if strings.Contains(line, "as-name:") || strings.Contains(line, "ASName:") {
					result = result + strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
				} else if strings.Contains(line, "descr:") {
					result = result + "\\n" + strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
				}
			}
			if result != "" {
				return fmt.Sprintf("AS%s\\n%s", asn, result)
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

	addPoint("Target: "+target, "[color=red,shape=diamond]")
	for serverID, server := range servers {
		response := responses[serverID]
		if len(response) == 0 {
			continue
		}
		addPoint(server, "[color=blue,shape=box]")
		// This is the best split point I can find for bird2
		routes := strings.Split(response, "unicast")

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
				// Try to parse the protocol instance name in Bird
				protocolNameRe := regexp.MustCompile(`\[(.*?) .*\]`)

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
