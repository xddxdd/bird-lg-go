package main

import (
	"encoding/json"
	"fmt"
	"net"
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

func graphvizEscape(s string) string {
	result, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	} else {
		return string(result)
	}
}

type ASNCache map[string]string

func (cache ASNCache) lookup(asn string) string {
	var representation string

	cachedValue, cacheOk := cache[asn]
	if cacheOk {
		return cachedValue
	}

	if setting.dnsInterface != "" {
		// get ASN representation using DNS
		records, err := net.LookupTXT(fmt.Sprintf("AS%s.%s", asn, setting.dnsInterface))
		if err == nil {
			result := strings.Join(records, " ")
			if resultSplit := strings.Split(result, " | "); len(resultSplit) > 1 {
				result = strings.Join(resultSplit[1:], "\n")
			}
			representation = fmt.Sprintf("AS%s\n%s", asn, result)
		}
	} else if setting.whoisServer != "" {
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
				representation = strings.Join(result, "\n")
			}
		}
	} else {
		representation = fmt.Sprintf("AS%s", asn)
	}

	cache[asn] = representation
	return representation
}

func birdRouteToGraphviz(servers []string, responses []string, targetName string) string {
	asnCache := make(ASNCache)
	graph := makeRouteGraph()

	makeEdgeAttrs := func(preferred bool) RouteAttrs {
		result := RouteAttrs{
			"fontsize": "12.0",
		}
		if preferred {
			result["color"] = "red"
		}
		return result
	}
	makePointAttrs := func(preferred bool) RouteAttrs {
		result := RouteAttrs{}
		if preferred {
			result["color"] = "red"
		}
		return result
	}

	target := "Target: " + targetName
	graph.AddPoint(target, RouteAttrs{"color": "red", "shape": "diamond"})

	for serverID, server := range servers {
		response := responses[serverID]
		if len(response) == 0 {
			continue
		}
		graph.AddPoint(server, RouteAttrs{"color": "blue", "shape": "box"})
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
				if i == 0 {
					src = server
				} else {
					src = asnCache.lookup(paths[i-1])
				}
				dst := asnCache.lookup(paths[i])

				graph.AddEdge(src, dst, strings.TrimSpace(protocolName+"\n"+via), makeEdgeAttrs(routePreferred))
				// Only set color for next step, origin color is set to blue above
				graph.AddPoint(dst, makePointAttrs(routePreferred))
			}

			// Last AS to destination
			src := asnCache.lookup(paths[len(paths)-1])
			graph.AddEdge(src, target, "", makeEdgeAttrs(routePreferred))
		}
	}

	return graph.ToGraphviz()
}

type RouteGraph struct {
	points map[string]RouteAttrs
	edges  map[RouteEdge]RouteAttrs
}
type RouteEdge struct {
	src   string
	dest  string
	label string
}
type RouteAttrs map[string]string

func attrsToString(attrs RouteAttrs) string {
	if len(attrs) == 0 {
		return ""
	}

	result := ""
	isFirst := true
	for k, v := range attrs {
		if isFirst {
			isFirst = false
		} else {
			result += ","
		}
		result += graphvizEscape(k) + "=" + graphvizEscape(v) + ""
	}

	return "[" + result + "]"
}

func makeRouteGraph() RouteGraph {
	return RouteGraph{
		points: make(map[string]RouteAttrs),
		edges:  make(map[RouteEdge]RouteAttrs),
	}
}

func (graph *RouteGraph) AddEdge(src string, dest string, label string, attrs RouteAttrs) {
	// Add edges with same src/dest separately, multiple edges with same src/dest could exist
	edge := RouteEdge{
		src:   src,
		dest:  dest,
		label: label,
	}

	_, exists := graph.edges[edge]
	if !exists {
		graph.edges[edge] = make(RouteAttrs)
	}

	for k, v := range attrs {
		graph.edges[edge][k] = v
	}
}

func (graph *RouteGraph) AddPoint(name string, attrs RouteAttrs) {
	graph.points[name] = attrs
}

func (graph *RouteGraph) ToGraphviz() string {
	var result string
	for name, attrs := range graph.points {
		result += fmt.Sprintf("%s %s;\n", graphvizEscape(name), attrsToString(attrs))
	}
	for edge, attrs := range graph.edges {
		attrsCopy := attrs
		if attrsCopy == nil {
			attrsCopy = make(RouteAttrs)
		}
		if len(edge.label) > 0 {
			attrsCopy["label"] = edge.label
		}
		result += fmt.Sprintf("%s -> %s %s;\n", graphvizEscape(edge.src), graphvizEscape(edge.dest), attrsToString(attrsCopy))
	}
	return "digraph {\n" + result + "}\n"
}
