package main

import (
	"bufio"
	"fmt"
	"lua-interpreter/internal/bytecode"
	"lua-interpreter/internal/lexer"
	"os"

	"lua-interpreter/internal/vm"
)

func main() {
	virtualMachine := vm.New()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("LuaGo VM (type 'exit' to quit)")
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
		code, err := bytecode.Compile(tokens)
		if err != nil {
			fmt.Println("Compile error:", err)
			continue
		}
		err = virtualMachine.Run(code)
		if err != nil {
			fmt.Println("Runtime error:", err)
		}
	}
}
