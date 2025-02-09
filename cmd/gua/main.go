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
			p := parser.NewParser(l)

			block, err := p.Parse()
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error parsing file: %s", err.Error()), -3)
			}

			val, err := block.Eval(ast.NewRootContext())
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error during evaluation: %s", err.Error()), -4)
			}

			fmt.Printf("Evaluation Result: %+v", val)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
