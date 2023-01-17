package imports

import (
	"os"
	"sync"
	"testing"
)

const inputFileContents = `package main

import (
	tf "thirdy.io/twofer"
	"example.com/exampkg"
	"github.com/random"
	"thirdy.io/two"
	t1 "github.com/thirdy.one"
	"os"
	"k8s.io/klog/v2"
)

func main() {
	os.Exit(86)
}
`

const expectedFileContents = `package main

import (
	"os"

	"github.com/random"

	"k8s.io/klog/v2"

	"thirdy.io/two"
	tf "thirdy.io/twofer"

	t1 "github.com/thirdy.one"

	"example.com/exampkg"
)

func main() {
	os.Exit(86)
}
`

const testFileName = "example.go"

func TestFormat(t *testing.T) {
	testDir, err := os.MkdirTemp("", "tools-test")
	if err != nil {
		t.Errorf("Failed to make temporary directory: %s", err)
	}
	defer os.RemoveAll(testDir)
	originalWD, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to read current working directory: %s", err)
	}
	defer os.Chdir(originalWD)
	os.Chdir(testDir)
	file, err := os.Create(testFileName)
	if err != nil {
		t.Errorf("Failed to create test input file: %s", err)
	}
	_, err = file.WriteString(inputFileContents)
	if err != nil {
		t.Errorf("Failed to write input file contents: %s", err)
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Failed to close input file: %s", err)
	}
	filesChan := make(chan string)
	exampleModule := "example.com/exampkg"
	dry := false
	list := false
	var wg sync.WaitGroup
	wg.Add(1)
	go Format(filesChan, &wg, []string{"thirdy.io/two", "github.com/thirdy.one"}, &exampleModule, &dry, &list)
	filesChan <- testFileName
	close(filesChan)
	wg.Wait()
	resultBytes, err := os.ReadFile(testFileName)
	if err != nil {
		t.Errorf("Failed to read test file: %s", err)
	}
	resultString := string(resultBytes)
	if resultString != expectedFileContents {
		t.Errorf("Expected %s but got %s", expectedFileContents, resultString)
	}
}
