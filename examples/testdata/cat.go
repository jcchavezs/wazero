package main

import (
	"os"
)

// main is the same as cat: "concatenate and print files."
func main() {
	// Start at arg[1] because args[0] is the program name.
	for i := 1; i < len(os.Args); i++ {
		bytes, err := os.ReadFile(os.Args[i])
		if err != nil {
			os.Exit(1)
		}

		// Use write to avoid needing to worry about Windows newlines.
		os.Stdout.Write(bytes)
	}
}
