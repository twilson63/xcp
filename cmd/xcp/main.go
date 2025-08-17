package main

import (
	"fmt"
	"os"
	"xcp/internal/cli"
)

func main() {
	opts := cli.Options{
		Args:   os.Args[1:],
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	c := cli.New(opts)
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
