package v1

import "regexp"

type ImportRegexp struct {
	Bucket string
	Regexp *regexp.Regexp
}
