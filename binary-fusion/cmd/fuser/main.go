package main

import (
	"fmt"
	"os"

	"binary-fusion/internal/elfcheck"
)

func main() {
	if len(os.Args) == 3 && os.Args[1] == "--sections" {
		if err := elfcheck.ListSections(os.Args[2]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) == 3 && os.Args[1] == "--segments" {
		if err := elfcheck.ListSegments(os.Args[2]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) == 5 && os.Args[1] == "--extract" {
		if err := elfcheck.ExtractSection(
			os.Args[2],
			os.Args[3],
			os.Args[4],
		); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	switch len(os.Args) {
	case 2:
		if err := elfcheck.Analyze(os.Args[1]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case 3:
		if err := elfcheck.CheckCompatibility(os.Args[1], os.Args[2]); err != nil {
			fmt.Println("Incompatible:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Usage:")
		fmt.Println("  fuser <binary>")
		fmt.Println("  fuser <binary1> <binary2>")
		fmt.Println("  fuser --sections <binary>")
		fmt.Println("  fuser --segments <binary>")
		fmt.Println("  fuser --extract <binary> <section> <output>")
		os.Exit(1)
	}
}
