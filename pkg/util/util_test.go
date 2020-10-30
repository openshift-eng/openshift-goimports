package util

import (
	"fmt"
	"os"
	"testing"
)

func TestIsGoFile(t *testing.T) {
	tests := []struct {
		filename string
		isGoFile bool
	}{
		{
			"test.go",
			true,
		},
		{
			"test.txt",
			false,
		},
	}

	for _, test := range tests {
		f, err := os.Create(fmt.Sprintf("%s/%s", os.TempDir(), test.filename))
		if err != nil {
			t.Fatalf("TempFile %s: %s", test.filename, err)
		}

		fi, err := f.Stat()
		if err != nil {
			t.Fatalf("TempFile %s: %s", test.filename, err)
		}

		if IsGoFile(fi) != test.isGoFile {
			t.Fatalf("TempFile %s: wanted %t, got %t, details: %#v", test.filename, test.isGoFile, IsGoFile(fi), fi)
		}
	}
}
