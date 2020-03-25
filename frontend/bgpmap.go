package main

import (
	"sort"
	"strings"
)

func birdRouteToGraphviz(servers []string, responses []string, target string) string {
	var edges string
	edges += "\"Target: " + target + "\" [color=red,shape=diamond];\n"
	for serverID, server := range servers {
		response := responses[serverID]
		if len(response) == 0 {
			continue
		}
		edges += "\"" + server + "\" [color=blue,shape=box];\n"
		routes := strings.Split(response, "\tvia ")
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
			if len(routeNexthop) == 0 || len(routeASPath) == 0 {
				continue
			}

			// Connect each node on AS path
			paths := strings.Split(strings.TrimSpace(routeASPath), " ")

			// First step starting from originating server
			if len(paths) > 0 {
				if len(routeNexthop) > 0 {
					// Edge from originating server to nexthop
					edges += "\"" + server + "\" -> \"Nexthop:\\n" + routeNexthop + "\"" + (map[bool]string{true: " [color=red]"})[routePreferred] + ";\n"
					// and from nexthop to AS
					edges += "\"Nexthop:\\n" + routeNexthop + "\" -> \"AS" + paths[0] + "\"" + (map[bool]string{true: " [color=red]"})[routePreferred] + ";\n"
					edges += "\"Nexthop:\\n" + routeNexthop + "\" [shape=diamond];\n"
				} else {
					// Edge from originating server to AS
					edges += "\"" + server + "\" -> \"AS" + paths[0] + "\"" + (map[bool]string{true: " [color=red]"})[routePreferred] + ";\n"
				}
			}

			// Following steps, edges between AS
			for pathIndex := range paths {
				if pathIndex == 0 {
					continue
				}
				edges += "\"AS" + paths[pathIndex-1] + "\" -> \"AS" + paths[pathIndex] + "\"" + (map[bool]string{true: " [color=red]"})[routePreferred] + ";\n"
			}
			// Last AS to destination
			edges += "\"AS" + paths[len(paths)-1] + "\" -> \"Target: " + target + "\"" + (map[bool]string{true: " [color=red]"})[routePreferred] + ";\n"
		}
		if !strings.Contains(edges, "\""+server+"\" ->") {
			// Cannot get path information from bird
			edges += "\"" + server + "\" -> \"Target: " + target + "\" [color=gray,label=\"?\"]"
		}
	}
	// Deduplication of edges: sort, then remove if current entry is prefix of next entry
	var result string
	edgesSorted := strings.Split(edges, ";\n")
	sort.Strings(edgesSorted)
	for edgeIndex, edge := range edgesSorted {
		if edgeIndex >= len(edgesSorted)-1 || !strings.HasPrefix(edgesSorted[edgeIndex+1], edge) {
			result += edge + ";\n"
		}
	}

	return "digraph {\n" + result + "}\n"
}
