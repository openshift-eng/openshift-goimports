package imports

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"sort"
	"testing"
)

func humanizeImportSpec(importSpecSlice []ast.ImportSpec) []string {
	var humanized []string
	for _, importSpec := range importSpecSlice {
		var spec string
		if importSpec.Name != nil {
			spec = importSpec.Name.Name
		}
		if importSpec.Path != nil {
			spec = fmt.Sprintf("%s %s", spec, importSpec.Path.Value)
		}
		humanized = append(humanized, spec)
	}

	return humanized
}

// TestByPathValue tests sorting imports by path value, ignoring the name for named imports
func TestByPathValue(t *testing.T) {
	tests := []struct {
		name string
		have []ast.ImportSpec
		want []ast.ImportSpec
	}{
		{
			name: "basic test",
			have: []ast.ImportSpec{
				ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/abc",
					},
				},
				ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/cba",
					},
				},
			},
			want: []ast.ImportSpec{
				ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/abc",
					},
				},
				ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/cba",
					},
				},
			},
		},
		{
			name: "basic test with named imports",
			have: []ast.ImportSpec{
				ast.ImportSpec{
					Name: &ast.Ident{
						Name: "cba",
					},
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/abc",
					},
				},
				ast.ImportSpec{
					Name: &ast.Ident{
						Name: "abc",
					},
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/cba",
					},
				},
			},
			want: []ast.ImportSpec{
				ast.ImportSpec{
					Name: &ast.Ident{
						Name: "cba",
					},
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/abc",
					},
				},
				ast.ImportSpec{
					Name: &ast.Ident{
						Name: "abc",
					},
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "github.com/example/cba",
					},
				},
			},
		},
	}

	for _, test := range tests {
		sort.Sort(byPathValue(test.have))
		if !reflect.DeepEqual(test.have, test.want) {
			t.Fatalf("test: %s, wanted: %#v, got %#v", test.name, humanizeImportSpec(test.want), humanizeImportSpec(test.have))
		}
	}
}
