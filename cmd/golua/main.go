package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"lua-interpreter/internal/lexer"
)

func main() {
	app := &cli.App{
		Name:  "golua",
		Usage: "GoLua - Lua Interpreter written in Go",
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
			var tokens []lexer.Token
			for l.HasNext() {
				token := l.NextToken()
				tokens = append(tokens, token)
			}
			if l.NextToken().Type != lexer.TokenEOF {
				return cli.Exit("Lexer did not return EOF token", -3)
			}
			// Print tokens for debugging
			for _, token := range tokens {
				fmt.Printf("Token: %s, Type: %s\n", token.Value, token.Type)
			}
			// Parse tokens
			//node, err := parser.Parse(tokens)
			//if err != nil {
			//	return cli.Exit(fmt.Sprintf("Error parsing tokens: %s", err.Error()), -4)
			//}
			// Print AST for debugging
			//fmt.Printf("AST: %+v\n", node)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
