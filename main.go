package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	env := NewEnv()
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
		tokens := Lex(line)
		ast, err := Parse(tokens)
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
