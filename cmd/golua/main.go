package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "golua",
		Usage: "GoLua - Lua Interpreter written in Go",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit("Provide path to lua file", 1)
			}
			path := c.Args().Get(0)
			_, err := os.ReadFile(path)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error reading file: %s", err.Error()), 1)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
