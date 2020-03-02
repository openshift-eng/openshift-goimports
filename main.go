package main

import (
	"bufio"
	"bytes"
	"flag"
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
	"sync"
)

type ImportRegexp struct {
	Bucket string
	Regexp *regexp.Regexp
}

var (
	wg           sync.WaitGroup
	impLine      = regexp.MustCompile(`^\s+(?:[\w\.]+\s+)?"(.+)"`)
	vendor       = regexp.MustCompile(`vendor/`)
	importRegexp []ImportRegexp
	files        = make(chan string, 10)
	importOrder  = []string{
		"standard",
		"other",
		"kubernetes",
		"openshift",
		"module",
	}
	beginPath = os.Args[1]
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

func formatImports(files chan string) {
	defer wg.Done()
	for path := range files {
		if len(path) == 0 {
			continue
		}
		fmt.Println(fmt.Sprintf("Processing %s", path))
		importGroups := map[string][]ast.ImportSpec{
			"standard":   []ast.ImportSpec{},
			"other":      []ast.ImportSpec{},
			"kubernetes": []ast.ImportSpec{},
			"openshift":  []ast.ImportSpec{},
			"module":     []ast.ImportSpec{},
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
			if len(i.Path.Value) == 0 {
				continue
			}
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
				for _, group := range importOrder {
					sort.Sort(ByPathValue(importGroups[group]))
					for n := range importGroups[group] {
						importGroups[group][n].EndPos = 0
						importGroups[group][n].Path.ValuePos = 0
						if importGroups[group][n].Name != nil {
							importGroups[group][n].Name.NamePos = 0
						}
						gen.Specs = append(gen.Specs, &importGroups[group][n])
						if n == 0 && group != importOrder[0] {
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
		info, err := os.Stat(path)
		err = ioutil.WriteFile(path, out, info.Mode())
		if err != nil {
			fmt.Println(fmt.Sprintf("%#v", err))
		}

	}
}

func isGoFile(f os.FileInfo) bool {
	// ignore non-Go files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func main() {
	modulePtr := flag.String("module", "blah", "package name")
	flag.Parse()

	importRegexp = []ImportRegexp{
		{Bucket: "module", Regexp: regexp.MustCompile(*modulePtr)},
		{Bucket: "kubernetes", Regexp: regexp.MustCompile("k8s.io")},
		{Bucket: "openshift", Regexp: regexp.MustCompile("github.com/openshift")},
		{Bucket: "other", Regexp: regexp.MustCompile("[a-zA-Z0-9]+\\.[a-zA-Z0-9]+/")},
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go formatImports(files)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.Walk(".",
			func(path string, f os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if f.IsDir() && f.Name() == "vendor" {
					return filepath.SkipDir
				}
				if isGoFile(f) && !vendor.MatchString(path) {
					fmt.Println(fmt.Sprintf("Queueing %s", path))
					files <- path
				}
				return nil
			})
		if err != nil {
			log.Println(err)
		}
		close(files)
	}()

	wg.Wait()
}
