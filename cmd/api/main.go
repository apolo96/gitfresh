package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", "error: "+err.Error())
		os.Exit(1)
	}
}
