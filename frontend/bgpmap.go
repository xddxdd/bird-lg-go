package main

import (
	"regexp"
	"strings"
)

// The protocol name for each route (e.g. "ibgp_sea02") is encoded in the form:
//
//	unicast [ibgp_sea02 2021-08-27 from fd86:bad:11b7:1::1] * (100/1015) [i]
var protocolNameRe = regexp.MustCompile(`\[(.*?) .*\]`)

// Try to split the output into one chunk for each route.
// Possible values are defined at https://gitlab.nic.cz/labs/bird/-/blob/v2.0.8/nest/rt-attr.c#L81-87
var routeSplitRe = regexp.MustCompile("(unicast|blackhole|unreachable|prohibited)")

var routeViaRe = regexp.MustCompile(`(?m)^\t(via .*?)$`)
var routeASPathRe = regexp.MustCompile(`(?m)^\tBGP\.as_path: (.*?)$`)

func makeEdgeAttrs(preferred bool) RouteAttrs {
	result := RouteAttrs{
		"fontsize": "12.0",
	}
	if preferred {
		result["color"] = "red"
	}
	return result
}

func makePointAttrs(preferred bool) RouteAttrs {
	result := RouteAttrs{}
	if preferred {
		result["color"] = "red"
	}
	return result
}

func birdRouteToGraph(servers []string, responses []string, target string) RouteGraph {
	graph := makeRouteGraph()

	graph.AddPoint(target, false, RouteAttrs{"color": "red", "shape": "diamond"})

	for serverID, server := range servers {
		response := responses[serverID]
		if len(response) == 0 {
			continue
		}
		graph.AddPoint(server, false, RouteAttrs{"color": "blue", "shape": "box"})
		routes := routeSplitRe.Split(response, -1)

		for routeIndex, route := range routes {
			if routeIndex == 0 {
				continue
			}

			var via string
			var paths []string
			var routePreferred bool = strings.Contains(route, "*")
			// Track non-BGP routes in the output by their protocol name, but draw them altogether in one line
			// so that there are no conflicts in the edge label
			var protocolName string

			if match := routeViaRe.FindStringSubmatch(route); len(match) >= 2 {
				via = strings.TrimSpace(match[1])
			}

			if match := routeASPathRe.FindStringSubmatch(route); len(match) >= 2 {
				pathString := strings.TrimSpace(match[1])
				if len(pathString) > 0 {
					paths = strings.Split(strings.TrimSpace(match[1]), " ")
					for i := range paths {
						paths[i] = strings.TrimPrefix(paths[i], "(")
						paths[i] = strings.TrimSuffix(paths[i], ")")
					}
				}
			}

			if match := protocolNameRe.FindStringSubmatch(route); len(match) >= 2 {
				protocolName = strings.TrimSpace(match[1])
				if routePreferred {
					protocolName = protocolName + "*"
				}
			}

			if len(paths) == 0 {
				graph.AddEdge(server, target, strings.TrimSpace(protocolName+"\n"+via), makeEdgeAttrs(routePreferred))
				continue
			}

			// Edges between AS
			for i := range paths {
				var src string
				var label string
				// Only show nexthop information on the first hop
				if i == 0 {
					src = server
					label = strings.TrimSpace(protocolName + "\n" + via)
				} else {
					src = paths[i-1]
					label = ""
				}
				dst := paths[i]

				graph.AddEdge(src, dst, label, makeEdgeAttrs(routePreferred))
				// Only set color for next step, origin color is set to blue above
				graph.AddPoint(dst, true, makePointAttrs(routePreferred))
			}

			// Last AS to destination
			src := paths[len(paths)-1]
			graph.AddEdge(src, target, "", makeEdgeAttrs(routePreferred))
		}
	}

	return graph
}

func birdRouteToGraphviz(servers []string, responses []string, targetName string) string {
	graph := birdRouteToGraph(servers, responses, targetName)
	return graph.ToGraphviz()
}
