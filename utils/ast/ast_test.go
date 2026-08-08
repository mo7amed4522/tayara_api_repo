package ast

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/mo7amed4522/tayara_api/global"
)

func TestAst(t *testing.T) {
	filename := filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "plugin", "gva", "plugin.go")
	// This is GVA scaffolding-runner that parses a file living in an external autocode root
	// (e.g. the generator's own workspace). Where that tree is absent (CI, fresh clones) it has
	// nothing to parse, so skip rather than hard-fail the suite.
	if _, statErr := os.Stat(filename); statErr != nil {
		t.Skipf("skipping ast render test: autocode tree not present at %s", filename)
		return
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		t.Error(err)
		return
	}
	err = ast.Print(fileSet, file)
	if err != nil {
		t.Error(err)
		return
	}
	err = printer.Fprint(os.Stdout, token.NewFileSet(), file)
	if err != nil {
		panic(err)
	}

}
