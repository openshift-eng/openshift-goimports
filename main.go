package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ImportRegexp struct {
	Bucket string
	Regexp *regexp.Regexp
}

var (
	impLine = regexp.MustCompile(`^\s+(?:[\w\.]+\s+)?"(.+)"`)

	files = make(chan string)
)

type ByPathValue []ast.ImportSpec

func (a ByPathValue) Len() int           { return len(a) }
func (a ByPathValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPathValue) Less(i, j int) bool { return a[i].Path.Value < a[j].Path.Value }

// taken from https://github.com/golang/tools/blob/71482053b885ea3938876d1306ad8a1e4037f367/internal/imports/imports.go#L380
func addImportSpaces(r io.Reader, breaks []string) ([]byte, error) {
	var out bytes.Buffer
	in := bufio.NewReader(r)
	inImports := false
	done := false
	for {
		s, err := in.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if !inImports && !done && strings.HasPrefix(s, "import") {
			inImports = true
		}
		if inImports && (strings.HasPrefix(s, "var") ||
			strings.HasPrefix(s, "func") ||
			strings.HasPrefix(s, "const") ||
			strings.HasPrefix(s, "type")) {
			done = true
			inImports = false
		}
		if inImports && len(breaks) > 0 {
			if m := impLine.FindStringSubmatch(s); m != nil {
				if m[1] == breaks[0] {
					out.WriteByte('\n')
					breaks = breaks[1:]
				}
			}
		}

		fmt.Fprint(&out, s)
	}
	return out.Bytes(), nil
}

func formatImports(path string) {
	importGroups := map[string][]ast.ImportSpec{
		"standard":   []ast.ImportSpec{},
		"other":      []ast.ImportSpec{},
		"kubernetes": []ast.ImportSpec{},
		"openshift":  []ast.ImportSpec{},
		"module":     []ast.ImportSpec{},
	}
	importOrder := []string{
		"standard",
		"other",
		"kubernetes",
		"openshift",
		"module",
	}
	importRegexp := []ImportRegexp{
		{Bucket: "module", Regexp: regexp.MustCompile("github.com/openshift/cluster-image-registry-operator")},
		{Bucket: "kubernetes", Regexp: regexp.MustCompile("k8s.io")},
		{Bucket: "openshift", Regexp: regexp.MustCompile("github.com/openshift")},
		{Bucket: "other", Regexp: regexp.MustCompile("[a-zA-Z0-9]+\\.[a-zA-Z0-9]+/")},
	}
	var breaks []string
	fs := token.NewFileSet()
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	f, err := parser.ParseFile(fs, "", contents, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range f.Imports {
		found := false
		for _, r := range importRegexp {

			if r.Regexp.MatchString(i.Path.Value) {
				importGroups[r.Bucket] = append(importGroups[r.Bucket], *i)
				found = true
				break
			}
		}
		if !found {
			importGroups["standard"] = append(importGroups["standard"], *i)
		}
	}

	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if ok && gen.Tok == token.IMPORT {
			gen.Specs = []ast.Spec{}
			newPos := gen.Lparen
			for _, group := range importOrder {
				sort.Sort(ByPathValue(importGroups[group]))
				for n := range importGroups[group] {
					newPos = token.Pos(int(newPos) + n)
					importGroups[group][n].Path.ValuePos = newPos
					importGroups[group][n].EndPos = newPos
					if importGroups[group][n].Name != nil {
						importGroups[group][n].Name.NamePos = newPos
					}

					gen.Specs = append(gen.Specs, &importGroups[group][n])
					if n == 0 {
						newstr, err := strconv.Unquote(importGroups[group][n].Path.Value)
						if err != nil {
							fmt.Println(err)
						}
						breaks = append(breaks, newstr)
					}
				}

			}
		}

	}

	printerMode := printer.TabIndent

	printConfig := &printer.Config{Mode: printerMode, Tabwidth: 4}

	var buf bytes.Buffer
	err = printConfig.Fprint(&buf, fs, f)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	out, err := addImportSpaces(bytes.NewReader(buf.Bytes()), breaks)
	out, err = format.Source(out)
	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("%#v", err))
	}
}

func isGoFile(f os.FileInfo) bool {
	// ignore non-Go files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func main() {
	go func() {
		err := filepath.Walk("/home/cdaley/go/src/github.com/openshift/cluster-image-registry-operator/pkg",
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if isGoFile(info) {
					files <- path
				}
				return nil
			})
		if err != nil {
			log.Println(err)
		}
	}()

	for file := range files {
		fmt.Println(fmt.Sprintf("Processing %s", file))
		formatImports(file)
	}
	close(files)
}
