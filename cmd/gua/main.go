package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"lua-interpreter/internal/interpreter"
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

			_, err = interpreter.Eval(string(buf))
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error: %s", err.Error()), -3)
			}
			//fmt.Printf("Result: %+v", val)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
