package elfcheck

import (
	"debug/elf"
	"fmt"
	"os"
)

func ExtractSection(binaryPath, sectionName, outputPath string) error {
	file, err := elf.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("open ELF file: %w", err)
	}
	defer file.Close()

	section := file.Section(sectionName)
	if section == nil {
		return fmt.Errorf("section %q was not found", sectionName)
	}

	data, err := section.Data()
	if err != nil {
		return fmt.Errorf("read section %q: %w", sectionName, err)
	}

	if len(data) == 0 {
		return fmt.Errorf("section %q is empty", sectionName)
	}

	if err := os.WriteFile(outputPath, data, 0o755); err != nil {
		return fmt.Errorf("write extracted file: %w", err)
	}

	fmt.Println("Extracted section:", sectionName)
	fmt.Println("Source binary:", binaryPath)
	fmt.Println("Output file:", outputPath)
	fmt.Println("Extracted size:", len(data), "bytes")

	return nil
}
