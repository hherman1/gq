package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

//go:embed program.go.tmpl
var programTmpl string

//go:embed gq/gq.go
var prelude string

func main() {
	if err := run(); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

func run() error {
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
	return nil
}
