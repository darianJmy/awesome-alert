package database

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestRuntimeInitDoesNotMigrateOrSeed(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "database.go", nil, 0)
	if err != nil {
		t.Fatalf("parse database.go: %v", err)
	}

	var initBody *ast.BlockStmt
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "Init" {
			initBody = function.Body
			break
		}
	}
	if initBody == nil {
		t.Fatal("Init function not found")
	}

	forbidden := map[string]bool{
		"AutoMigrate": true,
	}
	ast.Inspect(initBody, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := ""
		switch function := call.Fun.(type) {
		case *ast.Ident:
			name = function.Name
		case *ast.SelectorExpr:
			name = function.Sel.Name
		}
		if forbidden[name] || strings.HasPrefix(name, "seed") || strings.HasPrefix(name, "migrate") || strings.HasPrefix(name, "removeLegacy") {
			t.Errorf("runtime Init must not call %s", name)
		}
		return true
	})
}

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

func TestMySQLLiteralEscapesSeedData(t *testing.T) {
	got := mysqlLiteral("运维\\中枢's\n")
	want := "'运维\\\\中枢\\'s\\n'"
	if got != want {
		t.Fatalf("mysqlLiteral() = %q, want %q", got, want)
	}
}

func TestQuoteIdentifierEscapesBackticks(t *testing.T) {
	if got, want := quoteIdentifier("schema`version"), "`schema``version`"; got != want {
		t.Fatalf("quoteIdentifier() = %q, want %q", got, want)
	}
}
