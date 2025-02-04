package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
	"lua-interpreter/internal/parser"
)

func main() {
	app := &cli.App{
		Name:  "gua",
		Usage: "Gua - Lua Interpreter written in Go",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit("Provide path to lua file", -1)
			}
			path := c.Args().Get(0)
			buf, err := os.ReadFile(path)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error reading file: %s", err.Error()), -2)
			}
			l := lexer.NewLexer(string(buf))
			//var tokens []lexer.Token
			//for l.HasNext() {
			//	token := l.NextToken()
			//	tokens = append(tokens, token)
			//}
			//// Print tokens for debugging
			//for _, token := range tokens {
			//	fmt.Printf("Token: %s, Type: %s\n", token.Value, token.Type)
			//}

			p := parser.NewParser(l)
			block, err := p.Parse()
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error parsing file: %s", err.Error()), -3)
			}
			//// for debugging
			//fmt.Printf("Parsed Block: %+v\n", block)
			//for _, statement := range block.Statements {
			//	fmt.Printf("Statement: %+v\n", statement)
			//	if stmt, ok := statement.(*ast.BinaryOperatorExpression); ok {
			//		fmt.Printf("Binary Operator: %s, Left: %s, Right: %s\n", stmt.Operator.Value, stmt.Left, stmt.Right)
			//	}
			//}
			val := block.Eval(ast.NewRootContext())

			fmt.Printf("Evaluation Result: %+v", val)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
