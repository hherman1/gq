// Package gq provides functionality for use in gq's generated programs, and serves as the prelude to all of those programs.
package gq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// The central datastructure. If a Node is returned by a gq program it is pretty printed as JSON.
type Node struct {
	// Parsed JSON as a Go structure
	val interface{}
	// Any errors generated during construction or manipulation.
	err error
	// The command that generated this node.
	origin string
}

// Pretty prints the contained json. If something is wrong with the node, returns an error string instead.
func (n Node) String() string {
	if n.err != nil {
		return fmt.Sprintf("Error: %v: %v", n.origin, n.err)
	}
	var bs bytes.Buffer
	enc := json.NewEncoder(&bs)
	enc.SetIndent("", "\t")
	err := enc.Encode(n.val)
	if err != nil {
		return fmt.Sprintf("Error: %v: %v\n%v", n.origin, err, n.val)
	}
	return bs.String()
}

func (n *Node) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &n.val)
	if err != nil {
		return fmt.Errorf("gq.Node: %w", err)
	}
	return nil
}

// Fetches the current value of the node as an integer, if possible. Otherwise, sets the error for the node.
func (n *Node) Int() int {
	if f, ok := n.val.(float64); ok {
		return int(f)
	}
	n.trace("Int")
	n.err = fmt.Errorf("expected numeric, was %T", n.val)
	return 0
}

// Fetches the current value of the node as float, if possible. Otherwise, sets the error for the node.
func (n *Node) Float() float64 {
	if f, ok := n.val.(float64); ok {
		return f
	}
	n.trace("Float")
	n.err = fmt.Errorf("expected numeric, was %T", n.val)
	return 0
}

// Fetches the current value of the node as a map, if possible. Otherwise, sets the error for the node.
func (n *Node) MapValue() map[string]interface{} {
	if f, ok := n.val.(map[string]interface{}); ok {
		return f
	}
	n.trace("MapValue")
	n.err = fmt.Errorf("expected map, was %T", n.val)
	return map[string]interface{}{}
}

// Fetches the current value of the node as an array, if possible. Otherwise, sets the error for the node.
func (n *Node) Array() []interface{} {
	if f, ok := n.val.([]interface{}); ok {
		return f
	}
	n.trace("Array")
	n.err = fmt.Errorf("expected array, was %T", n.val)
	return []interface{}{}
}

// Fetches the current value of the node as string, if possible. Otherwise, sets the error for the node.
func (n *Node) Str() string {
	if s, ok := n.val.(string); ok {
		return s
	}
	n.trace("String")
	n.err = fmt.Errorf("expected string, was %T", n.val)
	return "error"
}

// G fetches the values at the given keys in the map node. If there is only one key, returns that key's value. If there are many keys, returns
// an array of their non-null values. If this is not a map node, returns an error node. If none of the keys are not found, returns null.
func (n *Node) G(keys ...string) *Node {
	if n.err != nil {
		return n
	}
	n.trace("G", keys)
	m, ok := n.val.(map[string]interface{})
	if !ok {
		n.err = fmt.Errorf("expected JSON map, found: %T", n.val)
		return n
	}
	if len(keys) == 1 {
		n.val = m[keys[0]]
		return n
	}
	var vals []interface{}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			vals = append(vals, v)
		}
	}
	if len(vals) == 0 {
		n.val = nil
		return n
	}
	n.val = vals
	return n
}

// I fetches the value at the given array indices. If this is not an array node, returns an error node. If none of the indices are found, returns null.
// If there is only one index given, returns just that value. Otherwise returns an array of values.
func (n *Node) I(is ...int) *Node {
	if n.err != nil {
		return n
	}
	n.trace("I", is)
	a, ok := n.val.([]interface{})
	if !ok {
		n.err = fmt.Errorf("expected JSON array, found: %T", n.val)
		return n
	}
	if len(is) == 1 {
		i := is[0]
		if i < 0 || i >= len(a) {
			n.val = nil
			return n
		}
		n.val = a[i]
		return n
	}
	var vals []interface{}
	for _, i := range is {
		if i < 0 || i >= len(a) {
			continue
		}
		vals = append(vals, a[i])
	}
	if len(vals) == 0 {
		n.val = nil
		return n
	}
	n.val = vals
	return n
}

// Filter removes nodes from the interior of the given map or array node if they fail the filter function.
func (n *Node) Filter(f func(*Node) bool) *Node {
	if n.err != nil {
		return n
	}
	cn := *n // operate on a copy in case we need to revert.
	cn.trace("Filter", "func")
	// map implementation
	if m, ok := cn.val.(map[string]interface{}); ok {
		filtered := make(map[string]interface{})
		for k, v := range m {
			subN := cn
			subN.trace("G", k)
			subN.val = v
			success := f(&subN)
			if subN.err != nil {
				// uh oh, error.
				*n = subN
				return n
			}
			if success {
				filtered[k] = v
			}
		}
		cn.val = filtered
		*n = cn
		return n
	}

	// array implementation
	if a, ok := cn.val.([]interface{}); ok {
		filtered := make([]interface{}, 0)
		for i, v := range a {
			subN := cn
			subN.trace("I", i)
			subN.val = v
			success := f(&subN)
			if subN.err != nil {
				// uh oh, error.
				*n = subN
				return n
			}
			if success {
				filtered = append(filtered, v)
			}
		}
		cn.val = filtered
		*n = cn
		return n
	}

	// whoops
	cn.err = fmt.Errorf("expected map or array, was %T", n.val)
	*n = cn
	return n
}

// Map replaces nodes from the interior of the given map or array node with the output of the function.
func (n *Node) Map(f func(*Node) *Node) *Node {
	if n.err != nil {
		return n
	}
	cn := *n // operate on a copy in case we need to revert.
	cn.trace("Map", "func")
	// map implementation
	if m, ok := cn.val.(map[string]interface{}); ok {
		filtered := make(map[string]interface{})
		for k, v := range m {
			subN := cn
			subN.trace("G", k)
			subN.val = v
			replace := f(&subN)
			if subN.err != nil {
				// uh oh, error.
				*n = subN
				return n
			}
			filtered[k] = replace.val
		}
		cn.val = filtered
		*n = cn
		return n
	}

	// array implementation
	if a, ok := cn.val.([]interface{}); ok {
		filtered := make([]interface{}, 0)
		for i, v := range a {
			subN := cn
			subN.trace("I", i)
			subN.val = v
			replace := f(&subN)
			if subN.err != nil {
				// uh oh, error.
				*n = subN
				return n
			}
			filtered = append(filtered, replace.val)
		}
		cn.val = filtered
		*n = cn
		return n
	}

	// whoops
	cn.err = fmt.Errorf("expected map or array, was %T", n.val)
	*n = cn
	return n
}

// Checks if this is a map node
func (n Node) IsMap() bool {
	_, ok := n.val.(map[string]interface{})
	return ok
}

// Adds the given trace information to the node.
func (n *Node) trace(method string, args ...interface{}) {
	var sargs []string
	for _, a := range args {
		sargs = append(sargs, fmt.Sprint(a))
	}
	origin := fmt.Sprintf("%v(%v)", method, strings.Join(sargs, ", "))
	if n.origin != "" {
		origin = fmt.Sprintf("%v: %v", n.origin, origin)
	}
	n.origin = origin
}
