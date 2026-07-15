package elfcheck

import (
	"debug/elf"
	"fmt"
)

func Analyze(path string) error {
	f, err := elf.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Println("File:", path)
	fmt.Println("Class:", f.Class)
	fmt.Println("Type:", f.Type)
	fmt.Println("File Type:", describeFileType(f.Type))
	fmt.Println("Machine:", f.Machine)
	fmt.Printf("Entry: 0x%x\n", f.Entry)

	return nil
}

func describeFileType(t elf.Type) string {
	switch t {
	case elf.ET_EXEC:
		return "Executable file"
	case elf.ET_DYN:
		return "Shared object / PIE executable"
	case elf.ET_REL:
		return "Relocatable object file"
	default:
		return "Unknown or unsupported ELF type"
	}
}

func ListSections(path string) error {
	f, err := elf.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Println("Sections for:", path)
	fmt.Println()

	for _, section := range f.Sections {
		flags := ""

		if section.Flags&elf.SHF_ALLOC != 0 {
			flags += "R"
		}
		if section.Flags&elf.SHF_WRITE != 0 {
			flags += "W"
		}
		if section.Flags&elf.SHF_EXECINSTR != 0 {
			flags += "X"
		}

		fmt.Printf(
			"%-25s %-15s %-5s Size=%d\n",
			section.Name,
			section.Type,
			flags,
			section.Size,
		)
	}

	return nil
}

func ListSegments(path string) error {
	f, err := elf.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Println("Segments for:", path)
	fmt.Println()

	for _, prog := range f.Progs {
		flags := ""

		if prog.Flags&elf.PF_R != 0 {
			flags += "R"
		}
		if prog.Flags&elf.PF_W != 0 {
			flags += "W"
		}
		if prog.Flags&elf.PF_X != 0 {
			flags += "X"
		}

		fmt.Printf(
			"%-15s %-5s Offset=0x%x Size=%d\n",
			prog.Type,
			flags,
			prog.Off,
			prog.Filesz,
		)
	}

	return nil
}
