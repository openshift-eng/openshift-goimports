# openshift-goimports
Organizes Go imports according to OpenShift best practices

## Summary
Organizes Go imports into the following groups:
 - **standard** - Any of the Go standard library packages
 - **other** - Anything not specifically called out in this list
 - **kubernetes** - Anything that starts with `k8s.io`
 - **openshift** - Anything that starts with `github.com/openshift`
 - **module** - Anything that is part of the current module

## Installation
```
# Install using go get
$ go get -u github.com/coreydaley/openshift-goimports
```

## Usage
```
Usage:
  openshift-goimports [flags]

Flags:
  -h, --help                             help for openshift-goimports
  -m, --module string                    The name of the go module. Example: github.com/example-org/example-repo (optional)
  -p, --path string                      The path to the go module to organize. Defaults to the current directory. (default ".") (optional)
  -v, --v Level                          number for the log level verbosity
```

## Examples
`openshift-goimports` will try to automatically determine the module using the `go.mod` file, if present, at the provided path location.

```
# Basic usage, command executed against current directory
$ openshift-goimports

# Basic usage with command executed in provided directory
$ openshift-goimports --module github.com/example-org/example-repo --path ~/go/src/example-org/example-repo
```
