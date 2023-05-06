package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RouteAttrs map[string]string

type RoutePoint struct {
	performLookup bool
	attrs         RouteAttrs
}

type RouteEdgeKey struct {
	src  string
	dest string
}

type RouteEdgeValue struct {
	label []string
	attrs RouteAttrs
}

type RouteGraph struct {
	points map[string]RoutePoint
	edges  map[RouteEdgeKey]RouteEdgeValue
}

func makeRouteGraph() RouteGraph {
	return RouteGraph{
		points: make(map[string]RoutePoint),
		edges:  make(map[RouteEdgeKey]RouteEdgeValue),
	}
}

func makeRoutePoint() RoutePoint {
	return RoutePoint{
		performLookup: false,
		attrs:         make(RouteAttrs),
	}
}

func makeRouteEdgeValue() RouteEdgeValue {
	return RouteEdgeValue{
		label: []string{},
		attrs: make(RouteAttrs),
	}
}

func (graph *RouteGraph) attrsToString(attrs RouteAttrs) string {
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
		result += graph.escape(k) + "=" + graph.escape(v) + ""
	}

	return "[" + result + "]"
}

func (graph *RouteGraph) escape(s string) string {
	result, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	} else {
		return string(result)
	}
}

func (graph *RouteGraph) AddEdge(src string, dest string, label string, attrs RouteAttrs) {
	// Add edges with same src/dest separately, multiple edges with same src/dest could exist
	edge := RouteEdgeKey{
		src:  src,
		dest: dest,
	}

	newValue, exists := graph.edges[edge]
	if !exists {
		newValue = makeRouteEdgeValue()
	}

	if len(label) != 0 {
		newValue.label = append(newValue.label, label)
	}
	for k, v := range attrs {
		newValue.attrs[k] = v
	}

	graph.edges[edge] = newValue
}

func (graph *RouteGraph) AddPoint(name string, performLookup bool, attrs RouteAttrs) {
	newValue, exists := graph.points[name]
	if !exists {
		newValue = makeRoutePoint()
	}

	newValue.performLookup = performLookup
	for k, v := range attrs {
		newValue.attrs[k] = v
	}

	graph.points[name] = newValue
}

func (graph *RouteGraph) GetEdge(src string, dest string) *RouteEdgeValue {
	key := RouteEdgeKey{
		src:  src,
		dest: dest,
	}
	value, ok := graph.edges[key]
	if ok {
		return &value
	} else {
		return nil
	}
}

func (graph *RouteGraph) GetPoint(name string) *RoutePoint {
	value, ok := graph.points[name]
	if ok {
		return &value
	} else {
		return nil
	}
}

func (graph *RouteGraph) ToGraphviz() string {
	var result string

	asnCache := make(ASNCache)

	for name, value := range graph.points {
		var representation string

		if value.performLookup {
			representation = asnCache.Lookup(name)
		} else {
			representation = name
		}

		attrsCopy := value.attrs
		if attrsCopy == nil {
			attrsCopy = make(RouteAttrs)
		}
		attrsCopy["label"] = representation

		result += fmt.Sprintf("%s %s;\n", graph.escape(name), graph.attrsToString(value.attrs))
	}

	for key, value := range graph.edges {
		attrsCopy := value.attrs
		if attrsCopy == nil {
			attrsCopy = make(RouteAttrs)
		}
		if len(value.label) > 0 {
			attrsCopy["label"] = strings.Join(value.label, "\n")
		}
		result += fmt.Sprintf("%s -> %s %s;\n", graph.escape(key.src), graph.escape(key.dest), graph.attrsToString(attrsCopy))
	}

	return "digraph {\n" + result + "}\n"
}
