package ast

import (
	"go/ast"
	"io"
)

type Ast interface {
	// Parse Parse file/code
	Parse(filename string, writer io.Writer) (file *ast.File, err error)
	// Rollback Rollback
	Rollback(file *ast.File) error
	// Injection Injection
	Injection(file *ast.File) error
	// Format Format output
	Format(filename string, writer io.Writer, file *ast.File) error
}
