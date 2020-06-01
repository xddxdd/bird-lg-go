package main

import (
	"fmt"
	"net"
	"strings"
)

func getASNRepresentation(asn string) string {
	records, err := net.LookupTXT(fmt.Sprintf("AS%s.%s", asn, setting.dnsInterface))
	if err != nil {
		// DNS query failed, only use ASN as output
		return fmt.Sprintf("AS%s", asn)
	}

	result := strings.Join(records, " ")
	if resultSplit := strings.Split(result, " | "); len(resultSplit) > 1 {
		result = strings.Join(resultSplit[1:], "\\n")
	}
	return fmt.Sprintf("AS%s\\n%s", asn, result)
}

func birdRouteToGraphviz(servers []string, responses []string, target string) string {
	graph := make(map[string]string)
	// Helper to add an edge
	addEdge := func(src string, dest string, attr string) {
		key := "\"" + src + "\" -> \"" + dest + "\""
		_, present := graph[key]
		// Do not remove edge's attributes if it's already present
		if present && len(attr) == 0 {
			return
		}
		graph[key] = attr
	}
	// Helper to set attribute for a point in graph
	addPoint := func(name string, attr string) {
		key := "\"" + name + "\""
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
		routes := strings.Split(response, "\tvia ")
		routeFound := false
		for routeIndex, route := range routes {
			var routeNexthop string
			var routeASPath string
			var routePreferred bool = routeIndex > 0 && strings.Contains(routes[routeIndex-1], "*")
			// Have to look at previous slice to determine if route is preferred, due to bad split point selection

			for _, routeParameter := range strings.Split(route, "\n") {
				if strings.HasPrefix(routeParameter, "\tBGP.next_hop: ") {
					routeNexthop = strings.TrimPrefix(routeParameter, "\tBGP.next_hop: ")
				} else if strings.HasPrefix(routeParameter, "\tBGP.as_path: ") {
					routeASPath = strings.TrimPrefix(routeParameter, "\tBGP.as_path: ")
				}
			}
			if len(routeASPath) == 0 {
				// Either this is not a BGP route, or the information is incomplete
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
				if len(routeNexthop) > 0 {
					// Edge from originating server to nexthop
					addEdge(server, "Nexthop:\\n"+routeNexthop, (map[bool]string{true: "[color=red]"})[routePreferred])
					// and from nexthop to AS
					addEdge("Nexthop:\\n"+routeNexthop, getASNRepresentation(paths[0]), (map[bool]string{true: "[color=red]"})[routePreferred])
					addPoint("Nexthop:\\n"+routeNexthop, "[shape=diamond]")
					routeFound = true
				} else {
					// Edge from originating server to AS
					addEdge(server, getASNRepresentation(paths[0]), (map[bool]string{true: "[color=red]"})[routePreferred])
					routeFound = true
				}
			}

			// Following steps, edges between AS
			for pathIndex := range paths {
				if pathIndex == 0 {
					continue
				}
				addEdge(getASNRepresentation(paths[pathIndex-1]), getASNRepresentation(paths[pathIndex]), (map[bool]string{true: "[color=red]"})[routePreferred])
			}
			// Last AS to destination
			addEdge(getASNRepresentation(paths[len(paths)-1]), "Target: "+target, (map[bool]string{true: "[color=red]"})[routePreferred])
		}

		if !routeFound {
			// Cannot find a path starting from this server
			addEdge(server, "Target: "+target, "[color=gray,label=\"?\"]")
		}
	}

	// Combine all graphviz commands
	var result string
	for edge, attr := range graph {
		result += edge + " " + attr + ";\n"
	}
	return "digraph {\n" + result + "}\n"
}
