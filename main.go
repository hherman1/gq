package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

//go:embed program.go.tmpl
var programTmpl string

//go:embed gq/gq.go
var prelude string

func main() {
	if err := run(); err != nil {
		fmt.Println("Error: ", err)
		fmt.Println(help)
		os.Exit(1)
	}
}

func run() error {
	// Check for help invocation.
	for _, a := range os.Args {
		if a == "-h" || a == "--help" {
			fmt.Println(help)
			return nil
		}
	}

	// prepare program
	tmpl, err := template.New("tmpl").Parse(programTmpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	program := strings.Join(os.Args[1:], " ")
	type Dot struct {
		Program string
		Prelude string
	}
	var runnable bytes.Buffer
	err = tmpl.Execute(&runnable, Dot{Program: program, Prelude: strings.ReplaceAll(prelude, "package gq", "")})
	if err != nil {
		return fmt.Errorf("generating program: %w", err)
	}

	// execute
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	p, err := i.Compile(runnable.String())
	if err != nil {
		return fmt.Errorf("compile generated program: %w", err)
	}
	_, err = i.Execute(p)
	if err != nil {
		return fmt.Errorf("execute program: %w", err)
	}
	/*
		// save to disk
		f, err := os.CreateTemp(os.TempDir(), "*.go")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()
		_, err = io.Copy(f, bytes.NewReader(runnable.Bytes()))
		if err != nil {
			return fmt.Errorf("write to temp file: %w", err)
		}
		err = f.Close()
		if err != nil {
			return fmt.Errorf("closing temp file: %w", err)
		}

		// execute new program
		cmd := exec.Command("go", "run", f.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			if cmd.ProcessState.ExitCode() == 1 {
				// the generated program compiled successfully, but had some error output. We should just exit immediately.
				os.Exit(1)
			}
			return fmt.Errorf("running generated program: %w", err)
		}
	*/
	return nil
}

const help = `gq executes go scripts against json.

Usage: cat f.json | gq 'j.G("hello").G("world")

The script you pass into 'gq' has access to a variable called 'j' which contains the parsed JSON. After your script is run, whatever remains in 'j' is printed. 'j' is of type *Node. The following functions are available to the script:

func (n *Node) Array() []interface{}
    Fetches the current value of the node as an array, if possible. Otherwise,
    sets the error for the node.

func (n *Node) Filter(f func(*Node) bool) *Node
    Filter removes nodes from the interior of the given map or array node if
    they fail the filter function.

func (n *Node) Float() float64
    Fetches the current value of the node as float, if possible. Otherwise, sets
    the error for the node.

func (n *Node) G(keys ...string) *Node
    G fetches the values at the given keys in the map node. If there is only one
    key, returns that key's value. If there are many keys, returns an array of
    their non-null values. If this is not a map node, returns an error node. If
    none of the keys are not found, returns null.

func (n *Node) I(is ...int) *Node
    I fetches the value at the given array indices. If this is not an array
    node, returns an error node. If none of the indices are found, returns null.
    If there is only one index given, returns just that value. Otherwise returns
    an array of values.

func (n *Node) Int() int
    Fetches the current value of the node as an integer, if possible. Otherwise,
    sets the error for the node.

func (n Node) IsMap() bool
    Checks if this is a map node

func (n *Node) Map(f func(*Node) *Node) *Node
    Map replaces nodes from the interior of the given map or array node with the
    output of the function.

func (n *Node) MapValue() map[string]interface{}
    Fetches the current value of the node as a map, if possible. Otherwise, sets
    the error for the node.

func (n *Node) Str() string
    Fetches the current value of the node as string, if possible. Otherwise,
    sets the error for the node.
`
