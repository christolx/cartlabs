package httpapi

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"github.com/christolx/cartlabs/internal/contract"
)

func TestRegisteredRoutesMatchOpenAPI(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "server.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	registered := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "HandleFunc" {
			return true
		}
		literal, ok := call.Args[0].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		pattern, err := strconv.Unquote(literal.Value)
		if err == nil && strings.Contains(pattern, "/api/v1/") {
			registered[pattern] = true
		}
		return true
	})

	spec, err := contract.GetSwagger()
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{}
	for path, item := range spec.Paths.Map() {
		for method := range item.Operations() {
			expected[strings.ToUpper(method)+" /api/v1"+path] = true
		}
	}
	for route := range expected {
		if !registered[route] {
			t.Errorf("OpenAPI route not registered: %s", route)
		}
	}
	for route := range registered {
		if !expected[route] {
			t.Errorf("registered route missing from OpenAPI: %s", route)
		}
	}
}
