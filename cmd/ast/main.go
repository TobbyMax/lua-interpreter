package main

import (
	"bufio"
	"fmt"
	"os"

	"lua-interpreter/internal/lexer"
	"lua-interpreter/internal/parser"
)

func main() {
	env := parser.NewEnv()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("LuaGo Interpreter (type 'exit' to quit)")
	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if line == "exit" {
			break
		}
		tokens := lexer.Lex(line)
		ast, err := parser.Parse(tokens)
		if err != nil {
			fmt.Println("Parse error:", err)
			continue
		}
		err = ast.Eval(env)
		if err != nil {
			fmt.Println("Runtime error:", err)
		}
	}
}
