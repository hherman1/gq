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

// G fetches the key at the given map node. If this is not a map node, returns an error node. If the key is not found, returns null.
func (n *Node) G(path string) *Node {
	if n.err != nil {
		return n
	}
	n.trace("G", path)
	m, ok := n.val.(map[string]interface{})
	if !ok {
		n.err = fmt.Errorf("expected JSON map, found: %T", n.val)
		return n
	}
	n.val = m[path]
	return n
}

// I fetches the value at the given array index. If this is not an array node, returns an error node. If the index is not found, returns null.
func (n *Node) I(i int) *Node {
	if n.err != nil {
		return n
	}
	n.trace("I", i)
	a, ok := n.val.([]interface{})
	if !ok {
		n.err = fmt.Errorf("expected JSON array, found: %T", n.val)
		return n
	}
	n.val = a[i]
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
