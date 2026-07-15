package main

import (
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
)

const selfPath = "/proc/self/exe"

func extractSection(sectionName string) (string, error) {
	file, err := elf.Open(selfPath)
	if err != nil {
		return "", fmt.Errorf("open running ELF: %w", err)
	}
	defer file.Close()

	section := file.Section(sectionName)
	if section == nil {
		return "", fmt.Errorf("section %q was not found", sectionName)
	}

	data, err := section.Data()
	if err != nil {
		return "", fmt.Errorf("read section %q: %w", sectionName, err)
	}

	tempFile, err := os.CreateTemp("", "binary-fusion-payload-*")
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}

	tempPath := tempFile.Name()

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("write payload: %w", err)
	}

	if err := tempFile.Chmod(0o700); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("set executable permission: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("close temporary file: %w", err)
	}

	return tempPath, nil
}

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
	julesPath, err := extractSection(".jules_payload")
	if err != nil {
		fmt.Println("Runtime loader error:", err)
		os.Exit(1)
	}
	defer os.Remove(julesPath)

	vincentPath, err := extractSection(".vincent_payload")
	if err != nil {
		fmt.Println("Runtime loader error:", err)
		os.Exit(1)
	}
	defer os.Remove(vincentPath)

	if err := runProgram(julesPath, false); err != nil {
		fmt.Println("Runtime loader error: Jules execution failed:", err)
		os.Exit(1)
	}

	if err := runProgram(vincentPath, true); err != nil {
		fmt.Println("Runtime loader error: Vincent execution failed:", err)
		os.Exit(1)
	}
}
