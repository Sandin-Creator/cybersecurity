package elfcheck

import (
	"debug/elf"
	"fmt"
)

func CheckCompatibility(file1, file2 string) error {
	f1, err := elf.Open(file1)
	if err != nil {
		return err
	}
	defer f1.Close()

	f2, err := elf.Open(file2)
	if err != nil {
		return err
	}
	defer f2.Close()

	fmt.Println("Comparing:")
	fmt.Println(" ", file1)
	fmt.Println(" ", file2)
	fmt.Println()

	if f1.Class != f2.Class {
		return fmt.Errorf("incompatible ELF class: %s vs %s", f1.Class, f2.Class)
	}

	if f1.Machine != f2.Machine {
		return fmt.Errorf("incompatible architecture: %s vs %s", f1.Machine, f2.Machine)
	}

	fmt.Println("Compatible: YES")
	fmt.Println("Class:", f1.Class)
	fmt.Println("Machine:", f1.Machine)

	return nil
}
