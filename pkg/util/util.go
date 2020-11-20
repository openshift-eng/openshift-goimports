/*
Copyright © 2020 Corey Daley <cdaley@redhat.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"os"
	"regexp"
	"strings"

	v1 "github.com/coreydaley/openshift-goimports/pkg/api/v1"
)

// IsGoFile returns whether or not a file is a Go source file
func IsGoFile(f os.FileInfo) bool {
	// ignore non-Go files
	return !f.IsDir() && !strings.HasPrefix(f.Name(), ".") && strings.HasSuffix(f.Name(), ".go")
}

// BuildImportRegexp builds the slice of ImportRegexp to use for sorting imports into buckets
func BuildImportRegexp(module string) []v1.ImportRegexp {
	return []v1.ImportRegexp{
		{Bucket: "module", Regexp: regexp.MustCompile(module)},
		{Bucket: "kubernetes", Regexp: regexp.MustCompile("k8s.io")},
		{Bucket: "openshift", Regexp: regexp.MustCompile("github.com/openshift")},
		{Bucket: "other", Regexp: regexp.MustCompile("[a-zA-Z0-9]+\\.[a-zA-Z0-9]+/")},
	}
}
