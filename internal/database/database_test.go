package database

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestFreshSchemaDoesNotRunLegacyDataConversions(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "database.go", nil, 0)
	if err != nil {
		t.Fatalf("parse database.go: %v", err)
	}
	var schemaBody *ast.BlockStmt
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "createCurrentSchema" {
			schemaBody = function.Body
			break
		}
	}
	if schemaBody == nil {
		t.Fatal("createCurrentSchema function not found")
	}
	ast.Inspect(schemaBody, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		identifier, ok := call.Fun.(*ast.Ident)
		if ok && (strings.HasPrefix(identifier.Name, "migrate") || strings.HasPrefix(identifier.Name, "removeLegacy")) {
			t.Errorf("fresh schema must not call %s", identifier.Name)
		}
		return true
	})
}

func TestSuffixedScriptNameKeepsDatabaseLimit(t *testing.T) {
	name := suffixedScriptName(strings.Repeat("脚", 100), 2)
	if got := len([]rune(name)); got != 100 {
		t.Fatalf("name length = %d, want 100", got)
	}
	if !strings.HasSuffix(name, " (2)") {
		t.Fatalf("name = %q, want duplicate suffix", name)
	}
}

func TestLegacyGitCredentialNameKeepsDatabaseLimit(t *testing.T) {
	name := legacyGitCredentialName(strings.Repeat("仓", 128), 12)
	if got := len([]rune(name)); got != 128 {
		t.Fatalf("name length = %d, want 128", got)
	}
	if !strings.HasSuffix(name, " 认证 (12)") {
		t.Fatalf("name = %q, want migration suffix", name)
	}
}
