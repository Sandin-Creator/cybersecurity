package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runProgram(path string, outputToStderr bool) error {
	cmd := exec.Command(path)

	if outputToStderr {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}

	return cmd.Run()
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: loader <first_program> <second_program>")
		os.Exit(1)
	}

	first := os.Args[1]
	second := os.Args[2]

	if err := runProgram(first, false); err != nil {
		fmt.Println("Error running first program:", err)
		os.Exit(1)
	}

	if err := runProgram(second, true); err != nil {
		fmt.Println("Error running second program:", err)
		os.Exit(1)
	}
}
