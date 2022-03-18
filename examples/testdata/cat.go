package main

import (
	"os"
)

// main is the same as cat.
func main() {
	for _, file := range os.Args {
		bytes, err := os.ReadFile(file)
		if err != nil {
			os.Exit(1)
		}

		// Use write to avoid needing to worry about Windows newlines.
		os.Stdout.Write(bytes)
	}
}
