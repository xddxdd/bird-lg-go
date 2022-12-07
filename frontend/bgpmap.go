package main

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
)

func graphvizEscape(s string) string {
	result, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	} else {
		return string(result)
	}
}

func getASNRepresentation(asn string) string {
	if setting.dnsInterface != "" {
		// get ASN representation using DNS
		records, err := net.LookupTXT(fmt.Sprintf("AS%s.%s", asn, setting.dnsInterface))
		if err == nil {
			result := strings.Join(records, " ")
			if resultSplit := strings.Split(result, " | "); len(resultSplit) > 1 {
				result = strings.Join(resultSplit[1:], "\n")
			}
			return fmt.Sprintf("AS%s\n%s", asn, result)
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
	graph := make(map[string](map[string]string))
	// Helper to add an edge
	addEdge := func(src string, dest string, attrKey string, attrValue string) {
		key := graphvizEscape(src) + " -> " + graphvizEscape(dest)
		_, present := graph[key]
		if !present {
			graph[key] = map[string]string{}
		}
		if attrKey != "" {
			graph[key][attrKey] = attrValue
		}
	}
	// Helper to set attribute for a point in graph
	addPoint := func(name string, attrKey string, attrValue string) {
		key := graphvizEscape(name)
		_, present := graph[key]
		if !present {
			graph[key] = map[string]string{}
		}
		if attrKey != "" {
			graph[key][attrKey] = attrValue
		}
	}
	// The protocol name for each route (e.g. "ibgp_sea02") is encoded in the form:
	//    unicast [ibgp_sea02 2021-08-27 from fd86:bad:11b7:1::1] * (100/1015) [i]
	protocolNameRe := regexp.MustCompile(`\[(.*?) .*\]`)
	// Try to split the output into one chunk for each route.
	// Possible values are defined at https://gitlab.nic.cz/labs/bird/-/blob/v2.0.8/nest/rt-attr.c#L81-87
	routeSplitRe := regexp.MustCompile("(unicast|blackhole|unreachable|prohibited)")

	addPoint("Target: "+target, "color", "red")
	addPoint("Target: "+target, "shape", "diamond")

	for serverID, server := range servers {
		response := responses[serverID]
		if len(response) == 0 {
			continue
		}
		addPoint(server, "color", "blue")
		addPoint(server, "shape", "box")
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
				edgeTarget := getASNRepresentation(paths[0])
				addEdge(server, edgeTarget, "fontsize", "12.0")
				if routePreferred {
					addEdge(server, edgeTarget, "color", "red")
					// Only set color for next step, origin color is set to blue above
					addPoint(edgeTarget, "color", "red")
				}
				if len(routeNexthop) > 0 {
					addEdge(server, edgeTarget, "label", protocolName + "\n" + routeNexthop)
				}
			}

			// Following steps, edges between AS
			for pathIndex := range paths {
				if pathIndex == 0 {
					continue
				}
				if routePreferred {
					addEdge(getASNRepresentation(paths[pathIndex-1]), getASNRepresentation(paths[pathIndex]), "color", "red")
					// Only set color for next step, origin color is set to blue above
					addPoint(getASNRepresentation(paths[pathIndex]), "color", "red")
				} else {
					addEdge(getASNRepresentation(paths[pathIndex-1]), getASNRepresentation(paths[pathIndex]), "", "")
				}
			}

			// Last AS to destination
			if routePreferred {
				addEdge(getASNRepresentation(paths[len(paths)-1]), targetNodeName, "color", "red")
			} else {
				addEdge(getASNRepresentation(paths[len(paths)-1]), targetNodeName, "", "")
			}
		}

		if len(nonBGPRoutes) > 0 {
			addEdge(server, targetNodeName, "label", strings.Join(nonBGPRoutes, "\n"))
			addEdge(server, targetNodeName, "fontsize", "12.0")

			if nonBGPRoutePreferred {
				addEdge(server, targetNodeName, "color", "red")
			}
		}
	}

	// Combine all graphviz commands
	var result string
	for edge, attr := range graph {
		result += edge;
		if len(attr) != 0 {
			result += " ["
			isFirst := true
			for k, v := range attr {
				if isFirst {
					isFirst = false
				} else {
					result += ","
				}
				result += graphvizEscape(k) + "=" + graphvizEscape(v) + "";
			}
			result += "]"
		}
		result += ";\n"
	}
	return "digraph {\n" + result + "}\n"
}
