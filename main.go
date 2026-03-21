package main

import (
	"os"

	"github.com/nattergabriel/reseed/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
