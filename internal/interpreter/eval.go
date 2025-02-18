package interpreter

import (
	"fmt"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
	"lua-interpreter/internal/parser"
)

func Eval(script string) (ast.Value, error) {
	l := lexer.NewLexer(script)
	p := parser.New(l)

	block, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("error during parsing: %w", err)
	}

	val, err := block.Eval(ast.NewRootContext())
	if err != nil {
		return nil, fmt.Errorf("error during evaluation: %w", err)
	}
	return val, nil
}
