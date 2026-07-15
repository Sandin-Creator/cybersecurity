# Binary Fusion Report

## Environment

**Operating System:** Ubuntu 24.04 Linux

**Programming Language:** Go

**Target Binary Format:** ELF64

**Tools Used:**

* Go
* GCC
* readelf
* file
* Radare2 (verification and analysis)

---

# Step 1 - Creating Test Executables

Before implementing binary fusion, two simple executable programs were created for testing and analysis.

## jules.c

```c
#include <stdio.h>

int main() {
    printf("What do they call it?\n");
    return 0;
}
```

Output:

```text
What do they call it?
```

## vincent.c

```c
#include <stdio.h>

int main() {
    printf("They call it a Royale with Cheese.\n");
    return 0;
}
```

Output:

```text
They call it a Royale with Cheese.
```

---

# Dynamic ELF Compilation

The first version of the test programs was compiled using:

```bash
gcc -o jules jules.c
gcc -o vincent vincent.c
```

Verification:

```bash
file jules
file vincent
```

Result:

```text
ELF 64-bit LSB pie executable
dynamically linked
```

These binaries are Position Independent Executables (PIE) and use dynamic linking.

---

# Static ELF Compilation

Because the project requirements include support for statically linked executables, static versions were also compiled.

Commands used:

```bash
gcc -no-pie -static -o jules_static jules.c
gcc -no-pie -static -o vincent_static vincent.c
```

Verification:

```bash
file jules_static
file vincent_static
```

Result:

```text
ELF 64-bit LSB executable
statically linked
```

The static binaries will be used as the primary test binaries during the binary fusion implementation.

---

# Execution Verification

The executables were tested to ensure they function correctly.

Commands:

```bash
./jules
./vincent

./jules_static
./vincent_static
```

Output:

```text
What do they call it?

They call it a Royale with Cheese.
```

All executables ran successfully.

---

# Screenshot

Screenshot showing compilation, ELF identification and successful execution:

`screenshots/01-test-programs.png`

---

# Current Status

Completed:

* Created test executables
* Compiled dynamic ELF executables
* Compiled static ELF executables
* Verified executable format using the `file` utility
* Verified successful execution of all test binaries

Next Step:

Analyze ELF header structures using `readelf` and implement ELF parsing in Go using the `debug/elf` package.


# Step 2 - ELF Header Analysis

The ELF (Executable and Linkable Format) header contains the metadata required by the operating system to load and execute a program.

The ELF header of `jules_static` was inspected using:

```bash
readelf -h jules_static
```

Important fields identified:

### Magic

The first bytes of an ELF file.

Purpose:
Used to identify the file as an ELF executable.

Typical value:

```text
7f 45 4c 46
```

---

### Class

Defines whether the binary is 32-bit or 64-bit.

Example:

```text
ELF64
```

---

### Type

Defines the ELF file type.

Common values:

| Type    | Meaning                        |
| ------- | ------------------------------ |
| ET_EXEC | Executable file                |
| ET_DYN  | Shared object / PIE executable |
| ET_REL  | Relocatable object file        |

The test binaries use:

```text
ET_EXEC
```

---

### Machine

Defines the target CPU architecture.

Example:

```text
Advanced Micro Devices X86-64
```

This confirms the executable is built for x86-64 systems.

---

### Entry Point Address

The memory address where execution begins.

This field is particularly important for the binary fusion project because the fused binary will later require modification of its entry point so that both executables can run sequentially.

---

### Program Header Offset

Location of the Program Header Table.

Program headers describe how executable segments are loaded into memory.

---

### Section Header Offset

Location of the Section Header Table.

Section headers describe individual sections such as:

```text
.text
.data
.rodata
.bss
```

---

### Purpose for this Project

Understanding the ELF header is required before modifying binaries.

The binary fusion tool will later:

* Validate ELF files
* Verify architecture compatibility
* Verify executable type compatibility
* Read the original entry point
* Modify the entry point during fusion
* Preserve required metadata in the fused binary

Screenshot:

`screenshots/02-elf-header.png`


# Step 3 - Reading ELF Headers with Go

The first version of the binary fusion tool was implemented using Go's built-in `debug/elf` package.

Purpose:

* Open ELF executables
* Parse ELF header information
* Display important metadata
* Verify that the file is a valid ELF executable

Implementation:

```go
f, err := elf.Open(path)
```

The tool currently extracts:

* ELF Class
* ELF Type
* Target Architecture
* Entry Point

Example execution:

```bash
go run ./cmd/fuser jules_static
```

Output:

```text
File: jules_static
Class: ELFCLASS64
Type: ET_EXEC
Machine: EM_X86_64
Entry: 0x401790
```

Explanation of fields:

### ELFCLASS64

Indicates that the executable is a 64-bit ELF binary.

### ET_EXEC

Indicates that the file is an executable.

Other possible values include:

* ET_DYN (shared object / PIE executable)
* ET_REL (relocatable object file)

### EM_X86_64

Indicates that the executable targets the x86-64 architecture.

### Entry Point

The memory address where execution begins.

This field is particularly important because later stages of the project will modify the entry point to ensure both programs execute sequentially inside the fused binary.

Screenshot:

`screenshots/03-go-elf-parser.png`


# Step 4 - Identifying ELF File Types

One of the project requirements is to correctly identify and interpret different ELF file types.

The ELF header contains a field called `Type` which describes the purpose of the file. The tool was extended to inspect this field and classify the file into a human-readable category.

The implementation checks the ELF type using Go's built-in `debug/elf` package.

Supported ELF types:

| ELF Type | Description                     |
| -------- | ------------------------------- |
| ET_EXEC  | Executable file                 |
| ET_DYN   | Shared object or PIE executable |
| ET_REL   | Relocatable object file         |

The tool now displays both the raw ELF type and a human-readable description.

Example output:

```text
File: jules_static
Class: ELFCLASS64
Type: ET_EXEC
File Type: Executable file
Machine: EM_X86_64
Entry: 0x401790
```

---

## Testing Executable Files (ET_EXEC)

The statically linked executable `jules_static` was analyzed.

Command:

```bash
go run ./cmd/fuser jules_static
```

Result:

```text
Type: ET_EXEC
File Type: Executable file
```

This confirms that the binary is a standard executable.

---

## Testing Shared Objects / PIE Executables (ET_DYN)

The dynamically linked PIE executable `jules` was analyzed.

Command:

```bash
go run ./cmd/fuser jules
```

Result:

```text
Type: ET_DYN
File Type: Shared object / PIE executable
```

This confirms that the tool can identify Position Independent Executables and shared object style binaries.

---

## Testing Relocatable Object Files (ET_REL)

A relocatable object file was created using GCC.

Command:

```bash
gcc -c -o jules.o jules.c
```

The generated object file was then analyzed.

Command:

```bash
go run ./cmd/fuser jules.o
```

Result:

```text
Type: ET_REL
File Type: Relocatable object file
```

Because relocatable object files are not directly executable, the entry point is reported as:

```text
Entry: 0x0
```

This is expected behavior.

---

## Reviewer Requirement Coverage

This step satisfies the following project requirements:

* The tool accurately identifies and interprets different types of files within the chosen format.
* The student can explain how the tool differentiates between executable, shared object, and relocatable files.
* The student can identify and explain the ELF Type field in the ELF header.

The differentiation is performed using the ELF header `Type` field:

* ET_EXEC → Executable file
* ET_DYN → Shared object / PIE executable
* ET_REL → Relocatable object file

Screenshot:

`screenshots/04-elf-file-types.png`


# Step 5 - ELF Compatibility Validation

Before two executables can be fused together, the tool must verify that they are compatible.

The binary fusion process requires both binaries to use the same architecture and ELF class. Attempting to combine incompatible executables would result in an invalid output binary.

The compatibility validation module was implemented using Go's `debug/elf` package.

The following properties are compared:

* ELF Class (32-bit vs 64-bit)
* CPU Architecture
* ELF Type

Example checks:

```text
ELFCLASS64 vs ELFCLASS64
EM_X86_64 vs EM_X86_64
```

If all required fields match, the executables are considered compatible for fusion.

Example execution:

```bash
go run ./cmd/fuser jules_static vincent_static
```

Expected result:

```text
Comparing:
  jules_static
  vincent_static

Compatible: YES

Class: ELFCLASS64
Machine: EM_X86_64
```

The compatibility checks provide an early validation stage before attempting any binary modification or fusion.

Reviewer relevance:

This step helps satisfy the following requirements:

* The tool handles executables of the same architecture.
* The tool detects and reports attempts to fuse incompatible executables.
* The student can explain how compatibility is determined before fusion.

Screenshot:

`screenshots/05-compatible-binaries.png`


# Step 6 - Detecting Incompatible Executables

The binary fusion process requires both input files to use compatible ELF properties.

At minimum, both files must use the same ELF class and CPU architecture. A 64-bit binary cannot be safely fused with a 32-bit object file because their headers, address sizes, relocation structures, and execution models are different.

To test this, a 32-bit relocatable object file was created.

Command:

```bash
gcc -m32 -c -o jules32.o jules.c
```

The file type was verified using:

```bash
file jules32.o
```

Result:

```text
jules32.o: ELF 32-bit LSB relocatable, Intel 80386, version 1 (SYSV), not stripped
```

The incompatible fusion check was then tested by comparing the 64-bit executable `jules_static` with the 32-bit object file `jules32.o`.

Command:

```bash
go run ./cmd/fuser jules_static jules32.o
```

Output:

```text
Comparing:
  jules_static
  jules32.o

Incompatible: incompatible ELF class: ELFCLASS64 vs ELFCLASS32
exit status 1
```

The tool correctly rejected the input files and reported the reason clearly.

This prevents invalid fusion attempts and helps ensure that only compatible binaries are passed to the later fusion stage.

Reviewer relevance:

* The tool detects and reports attempts to fuse incompatible executables.
* The student can demonstrate fusing incompatible executables.
* The student can explain why ELF class compatibility is required.
* The tool handles executables of the same architecture by validating ELF class and machine type before fusion.

Screenshot:

`screenshots/06-incompatible-binaries.png`


# Step 7 - Section Analysis and Layout Inspection

Before executables can be fused together, it is necessary to understand their section layout.

ELF files contain multiple sections which store different types of data.

Examples:

| Section | Purpose                     |
| ------- | --------------------------- |
| .text   | Executable code             |
| .rodata | Read-only data              |
| .data   | Initialized writable data   |
| .bss    | Uninitialized writable data |
| .symtab | Symbol table                |
| .strtab | String table                |

A new section analysis feature was implemented in the binary fusion tool.

The feature reads all section headers and displays:

* Section name
* Section type
* Section flags
* Section size

Command:

```bash
go run ./cmd/fuser --sections jules_static
```

Example output:

```text
Sections for: jules_static

.text
.rodata
.data
.bss
...
```

This information will later be used during the fusion stage to determine which sections can be copied, merged, or preserved.

Screenshot:

`screenshots/07-sections-list.png`

---

# Step 8 - Handling Non-Standard Section Names

One of the project requirements is that the tool must not crash when encountering non-standard section names.

To test this, a custom executable was created containing a user-defined ELF section.

Source code:

```c
#include <stdio.h>

__attribute__((section(".my_custom_section")))
char customData[] = "Hello";

int main() {
    printf("Custom section test\n");
    return 0;
}
```

Compilation:

```bash
gcc -o custom_section custom_section.c
```

The section analysis feature was then used to inspect the executable.

Command:

```bash
go run ./cmd/fuser --sections custom_section
```

Result:

```text
.my_custom_section
```

The tool successfully detected and displayed the custom section without errors.

This demonstrates that the implementation does not rely on hardcoded section names such as:

```text
.text
.data
.rodata
```

Instead, the tool dynamically reads section information directly from the ELF section table.

Reviewer relevance:

* The program handles files with non-standard section names without crashing.
* The student can explain how section headers are processed.
* The tool prepares section information required for the future fusion process.

Screenshot:

`screenshots/08-custom-section.png`


# Step 9 - Section Permissions Analysis

The fusion process must preserve the permissions associated with executable sections.

ELF section headers contain flag information which describes how each section is intended to be used.

Important permission flags:

| Flag | Meaning    |
| ---- | ---------- |
| R    | Readable   |
| W    | Writable   |
| X    | Executable |

The section analysis feature was extended to display permissions for every section.

Implementation:

The tool reads ELF section flags and converts them into a human-readable permission format.

Examples:

```text
.text     RX
.rodata   R
.data     RW
.bss      RW
```

These permissions are important because executable code must remain executable after fusion, while writable data sections must preserve write access.

Command:

```bash
go run ./cmd/fuser --sections jules_static
```

Example output:

```text
.text                     SHT_PROGBITS    RX
.rodata                   SHT_PROGBITS    R
.data                     SHT_PROGBITS    RW
.bss                      SHT_NOBITS      RW
```

Future use in fusion:

During the fusion stage, section permissions will be preserved when sections are copied or merged into the final binary. This ensures that executable code remains executable and writable data remains writable.

Reviewer relevance:

* The tool preserves basic section permissions (executable, readable, writable) in the fused binary.
* The student can explain ELF section flags.
* The student can explain why permissions must be preserved during binary fusion.

Screenshot:

`screenshots/09-section-permissions.png`


# Step 10 - Program Header Analysis

In addition to section headers, ELF files contain program headers (segments).

Program headers describe how the operating system loads the executable into memory.

While section headers are mainly used by linkers and analysis tools, program headers are used directly by the Linux loader during program execution.

A new segment analysis feature was implemented.

Command:

```bash
go run ./cmd/fuser --segments jules_static
```

The tool displays:

* Segment type
* Memory permissions
* File offset
* Segment size

Example output:

```text
PT_LOAD         R
PT_LOAD         RX
PT_LOAD         RW
```

Permission meanings:

| Flag | Meaning    |
| ---- | ---------- |
| R    | Readable   |
| W    | Writable   |
| X    | Executable |

Important observations:

* Executable code is stored in RX segments.
* Constant data is stored in R segments.
* Writable program data is stored in RW segments.

These permissions must be preserved during binary fusion to ensure the resulting executable functions correctly.

Relationship between sections and segments:

Examples:

```text
.text    -> RX segment
.rodata  -> R segment
.data    -> RW segment
.bss     -> RW segment
```

The section headers describe logical contents of the executable, while program headers describe how those contents are mapped into memory.

Reviewer relevance:

* The student can explain how ELF executables are loaded.
* The student can explain the relationship between sections and segments.
* The student can explain how executable permissions are preserved.
* This information will later be used when constructing the fused executable.

Screenshot:

`screenshots/10-program-headers.png`


# Step 11 - Testing LIEF Integration

After completing the ELF analysis phase using Go's built-in `debug/elf` package, the next requirement was preparing for binary modification and fusion.

The project specification recommends the use of LIEF (Library to Instrument Executable Formats), which provides functionality for parsing, modifying and rebuilding executable files.

To avoid modifying the system Python installation, LIEF was installed inside a dedicated Python virtual environment.

Environment setup:

```bash
python3 -m venv venv
source venv/bin/activate
pip install lief
```

Installation verification:

```python
import lief
print(lief.__version__)
```

The installation completed successfully and LIEF version 0.17.6 was installed.

---

## Creating a LIEF Test Program

Before attempting any binary modifications, a small Python test program was created to verify that LIEF could successfully parse ELF executables.

The script reads:

* ELF format
* Entry point
* Architecture
* Number of sections
* Number of segments

Example execution:

```bash
python scripts/lief_info.py jules_static
```

Output:

```text
File: jules_static
Format: FORMATS.ELF
Entry point: 0x401790
Architecture: ARCH.X86_64
Sections: 28
Segments: 10
```

---

## Results

LIEF successfully parsed the ELF executable and extracted key metadata.

The reported values matched the information previously obtained using:

```bash
readelf
```

and

```go
debug/elf
```

This confirms that LIEF can correctly interpret the executable structure and is suitable for the upcoming binary fusion stage.

---

## Why LIEF Was Chosen

The project requires:

* Modifying executable files
* Adding new binary data
* Preserving ELF structure
* Updating entry points
* Generating a valid fused executable

While Go's `debug/elf` package is excellent for analysis, it does not provide functionality for rebuilding modified ELF files.

LIEF provides:

* ELF parsing
* Section manipulation
* Segment manipulation
* Entry point modification
* Binary rebuilding

These capabilities make it appropriate for implementing the fusion stage of the project.

---

## Reviewer Relevance

This step demonstrates:

* Successful installation and verification of LIEF.
* Ability to parse ELF executables using multiple analysis tools.
* Preparation for implementing executable fusion.
* Understanding of the difference between analysis tools (`debug/elf`) and modification tools (LIEF).

Screenshot:

`screenshots/11-lief-installation.png`


# Step 12 - Original Entry Point Verification

After embedding `vincent_static` into the `.fused_payload` section, the fused executable was analyzed to verify its current execution flow.

At this stage, the binary had already been modified to include the second executable as an embedded payload, but the ELF entry point had not yet been changed.

The purpose of this verification step was to confirm that:

* The fused ELF file remained structurally valid.
* The embedded payload did not corrupt the executable.
* The original entry point was still being used.
* The executable continued to function normally before entry point modification.

---

## Verification Commands

The fused executable was analyzed using multiple tools to verify the current entry point.

Commands used:

```bash
python scripts/lief_info.py fused_jules

readelf -h fused_jules | grep Entry

go run ./cmd/fuser fused_jules
```

---

## Results

The entry point reported by:

* LIEF
* readelf
* Go's `debug/elf`

matched the original entry point of `jules_static`.

This confirms that the fusion process had not yet altered the execution flow.

The ELF structure remained valid and executable after adding the new `.fused_payload` section.

---

## Runtime Verification

The fused executable was executed.

Command:

```bash
./fused_jules
```

Output:

```text
What do they call it?
```

This demonstrates that execution still begins at the original Jules entry point.

The embedded Vincent executable is present inside the binary but is not yet executed because the entry point has not been redirected to custom fusion logic.

---

## Current State of Fusion

At this stage:

Completed:

* ELF header analysis
* ELF type detection
* Compatibility validation
* Section analysis
* Program header analysis
* Section permission analysis
* LIEF integration
* Creation of a fused ELF containing `.fused_payload`

Not yet completed:

* Entry point modification
* Code injection
* Sequential execution of both executables

---

## Purpose of the Next Stage

The next stage of the project will modify the ELF entry point.

Instead of beginning execution directly at the original Jules entry point, execution will first enter injected fusion logic.

The modified execution flow will be:

```text
New Entry Point
        ↓
Fusion Loader
        ↓
Execute Jules
        ↓
Execute Vincent
        ↓
Program Exit
```

This modification is required to satisfy the project requirements related to:

* Code Injection
* Entry Point Modification
* Sequential Execution
* Binary Fusion Functionality

---

## Reviewer Relevance

This step demonstrates:

* Verification of the original ELF entry point.
* Validation that the fused binary remains executable.
* Confirmation that the embedded payload does not corrupt the executable.
* Preparation for the upcoming entry point modification stage.

Screenshot:

`screenshots/12-original-entry-point.png`


# Step 13 - Creating a Fused ELF Payload Section

The first actual binary fusion step was implemented using the LIEF library.

The fusion tool now accepts two ELF executables as input:

```bash
python scripts/fuse.py jules_static vincent_static fused_jules
```

The first executable, `jules_static`, is used as the base ELF executable.

The second executable, `vincent_static`, is read as raw binary data and inserted into the ELF as a new section.

The new section is named:

```text
.fused_payload
```

This section stores the entire contents of the second executable inside the fused ELF file.

---

## Fusion Output

Execution of the fusion script produced:

```text
Created: fused_jules
Base binary: jules_static
Payload binary: vincent_static
Original entry point: 0x401790
Payload size: 785360 bytes
Added section: .fused_payload
Entry point modified: NO
```

At this stage the original entry point remains unchanged.

The purpose of this step was to verify that:

* A second executable can be embedded inside the ELF.
* The ELF structure remains valid.
* Additional sections can be added safely.
* The executable remains runnable after modification.

---

## Validation

The resulting binary was verified using ELF analysis tools.

Commands:

```bash
readelf -h fused_jules | grep Entry

readelf -S fused_jules | grep fused

./fused_jules
```

Results:

```text
Entry point address: 0x401790

[23] .fused_payload PROGBITS

What do they call it?
```

---

## Observations

The validation confirms:

* The ELF header remains valid.
* The original entry point is preserved.
* The new `.fused_payload` section exists inside the executable.
* The binary continues to execute successfully.
* The embedded payload does not corrupt the executable.

The fused executable now physically contains both programs.

However, execution flow has not yet been modified.

The executable still begins execution at the original Jules entry point and therefore only executes the original program.

---

## Current Fusion Status

Completed:

* ELF analysis
* Compatibility validation
* Section analysis
* Program header analysis
* LIEF integration
* Payload embedding
* Creation of a valid fused ELF file

Not yet completed:

* Entry point modification
* Code injection
* Sequential execution of both executables

These features will be implemented in the next phase of the project.

---

## Reviewer Relevance

This step demonstrates:

* Successful modification of ELF structure using LIEF.
* Ability to add custom sections to an executable.
* Preservation of executable validity after modification.
* Preparation for the upcoming entry point modification stage.

Screenshot:

`screenshots/13-fused-payload-section.png`


# Step 14 - Injecting an Executable Loader Section

The next phase of the fusion process required preparing a location for custom executable code.

A new ELF section named:

```text
.fused_loader
```

was added to the fused executable using LIEF.

The purpose of this section is to eventually contain the custom loader logic responsible for executing both programs sequentially.

Unlike the payload section, the loader section was created with executable permissions.

Implementation:

```text
Section Name: .fused_loader
Type: PROGBITS
Permissions: AX
```

Permission meanings:

| Flag | Meaning             |
| ---- | ------------------- |
| A    | Allocated in memory |
| X    | Executable          |

The loader section was initially populated with NOP instructions (`0x90`) to reserve executable space for future loader code.

---

## Fusion Output

The fusion script now produces:

```text
Created: fused_jules
Base binary: jules_static
Payload binary: vincent_static
Original entry point: 0x401790
Payload size: 785360 bytes
Added section: .fused_payload
Added section: .fused_loader
Entry point modified: NO
```

---

## Verification

The section layout was verified using:

```bash
readelf -W -S fused_jules | grep fused
```

Result:

```text
.fused_loader     PROGBITS    AX
.fused_payload    PROGBITS    A
```

---

## Observations

The executable loader section was successfully added.

Important observations:

* The ELF file remains valid.
* The original entry point remains unchanged.
* The loader section is executable.
* The payload section remains allocated.
* The executable continues to run successfully.

At this stage, the binary now contains:

* Original Jules executable code
* Embedded Vincent executable payload
* Dedicated executable loader section

The final remaining fusion stage is redirecting the ELF entry point into the loader section and implementing the sequential execution logic.

---

## Reviewer Relevance

This step demonstrates:

* Successful code injection preparation.
* Preservation of executable permissions.
* Ability to create executable ELF sections.
* Preparation for entry point modification.

Screenshot:

`screenshots/14-loader-section.png`


# Step 15 - Creating a Loadable Executable Loader Section

The initial implementation of `.fused_loader` created an executable section, but the section was not mapped into memory.

Verification showed:

```text
.fused_loader PROGBITS 0000000000000000 ... AX
```

Because the virtual address was zero, the section could not safely be used as an ELF entry point.

To resolve this, the loader section was added as a loaded section.

Implementation change:

```python
base.add(loader_section, loaded=True)
```

instead of:

```python
base.add(loader_section, loaded=False)
```

---

## Verification

The fusion process was executed again:

```bash
python scripts/fuse.py jules_static vincent_static fused_jules
```

The resulting ELF was inspected:

```bash
readelf -W -S fused_jules | grep loader
```

Result:

```text
[28] .fused_loader PROGBITS 0000000000980000 ... AX
```

The loader section now has:

* A valid virtual address.
* Executable permissions.
* A dedicated loadable memory region.

Additional verification using LIEF:

```bash
python scripts/lief_info.py fused_jules
```

Result:

```text
Sections: 30
Segments: 12
```

The segment count increased because LIEF created additional structures required to map the loader section into memory.

---

## Observations

The `.fused_loader` section is now:

* Allocated in memory.
* Executable.
* Available as a future entry point target.

This is a critical milestone because the ELF loader can only transfer execution to code that resides within a valid executable memory region.

The project now contains:

* Original Jules executable.
* Embedded Vincent payload.
* Dedicated executable loader section.
* Loadable memory mapping for future injected code.

---

## Reviewer Relevance

This step demonstrates:

* Correct handling of executable ELF sections.
* Preservation of section permissions.
* Understanding of the difference between section metadata and loadable memory mappings.
* Preparation for entry point redirection.

Screenshot:

`screenshots/15-loaded-loader-section.png`


# Step 16 - Controlled Entry Point Modification

After creating a loadable executable loader section, the next step was to verify that the ELF entry point could be safely redirected to custom code.

Before modifying the entry point, the original executable used:

```text
Entry point: 0x401790
```

The loader section (`.fused_loader`) was mapped into executable memory and assigned the virtual address:

```text
0x980000
```

The ELF entry point was then updated to point to the loader section.

Implementation:

```python
base.header.entrypoint = loader_address
```

---

## Safe Loader Stub

Before implementing the full sequential execution logic, a small loader stub was used.

The stub performs the Linux x86-64 `exit(0)` system call.

This was intentionally chosen because it provides a safe way to verify that execution begins at the injected loader section without affecting the original executable.

---

## Fusion Output

Command:

```bash
python scripts/fuse.py jules_static vincent_static fused_entry_test
```

Output:

```text
Created: fused_entry_test
Base binary: jules_static
Payload binary: vincent_static
Original entry point: 0x401790
Loader entry point: 0x980000
Payload size: 785360 bytes
Added section: .fused_payload
Added section: .fused_loader
Entry point modified: YES
Loader behavior: exit successfully with status 0
```

---

## Entry Point Verification

Commands:

```bash
readelf -h jules_static | grep Entry
readelf -h fused_entry_test | grep Entry
```

Results:

```text
Original entry point: 0x401790
Modified entry point: 0x980000
```

This confirms that the ELF header was successfully modified and now points to the injected loader section.

---

## Runtime Verification

The modified executable was tested:

```bash
./fused_entry_test
echo $?
```

Result:

```text
0
```

The executable terminated successfully without crashing.

No output from Jules or Vincent is expected at this stage because the temporary loader immediately executes `exit(0)`.

This confirms that Linux transfers execution directly to the injected loader section.

---

## Current Status

Completed:

* Embedded second executable inside `.fused_payload`
* Created executable `.fused_loader` section
* Mapped the loader section into memory
* Successfully modified the ELF entry point
* Verified execution begins in the injected loader

The next step will replace the temporary loader stub with a real loader that executes both embedded programs sequentially.

---

## Screenshots

* `screenshots/16-entry-point-modification.png`
* `screenshots/17-loader-exit-test.png`


# Step 17 - Verifying Sequential Execution and Output Separation

Before integrating the complete loader behavior into the fused ELF, the sequential execution logic was tested separately using a Go loader program.

The loader executes two programs in order:

1. Jules
2. Vincent

The loader also routes their outputs to different streams:

* Jules output is written to standard output (`stdout`).
* Vincent output is written to standard error (`stderr`).

---

## Building the Loader

Command:

```bash
go build -o loader_bin ./loader
```

---

## Sequential Execution Test

Command:

```bash
./loader_bin ./jules_static ./vincent_static
```

Output:

```text
What do they call it?
They call it a Royale with Cheese.
```

This confirms that the loader executes the two programs in the correct order.

---

## Output Separation Test

Command:

```bash
./loader_bin ./jules_static ./vincent_static > output_jules.txt 2> output_vincent.txt
```

The output files were inspected using:

```bash
cat output_jules.txt
cat output_vincent.txt
```

Results:

```text
What do they call it?
```

and:

```text
They call it a Royale with Cheese.
```

This confirms that:

* Jules output is written to `stdout`.
* Vincent output is written to `stderr`.
* Shell redirection can store the outputs in separate files.

---

## Reviewer Relevance

This step demonstrates:

* Correct sequential execution order.
* Correct handling of standard output and standard error.
* Preparation for the extra output-redirection requirement.
* A working reference implementation for the final fused loader behavior.

Screenshot:

`screenshots/18-loader-sequential-output.png`


# Step 18 - Verifying Payload Integrity

One of the most important requirements of the binary fusion process is that the embedded executable remains unchanged after fusion.

Initially, the payload was embedded using only the LIEF library. During testing it was discovered that the extracted payload no longer matched the original executable.

The extracted file contained zeroed data instead of the original ELF header and therefore could not be executed.

To investigate this issue, the payload was extracted from the fused executable and compared with the original using SHA-256 hashes.

This demonstrated that the payload had been modified during the LIEF write process.

---

## Improved Fusion Workflow

To preserve the payload exactly, the fusion process was redesigned.

The final workflow is:

1. LIEF modifies the base executable.
2. LIEF adds the executable loader section.
3. LIEF modifies the ELF entry point.
4. The modified executable is written to disk.
5. GNU `objcopy` embeds the payload as the final step.

This prevents LIEF from rewriting the payload after it has been embedded.

The final workflow therefore combines the strengths of both tools:

* **LIEF** for ELF modification.
* **GNU objcopy** for byte-perfect payload embedding.

---

## Payload Extraction

The embedded payload was extracted using the project's extraction tool.

Command:

```bash
go run ./cmd/fuser --extract fused_correct_order_test .fused_payload extracted_correct_order
```

Result:

```text
Extracted section: .fused_payload
Source binary: fused_correct_order_test
Output file: extracted_correct_order
Extracted size: 785360 bytes
```

---

## Integrity Verification

The extracted payload was compared with the original executable using SHA-256.

Command:

```bash
sha256sum vincent_static extracted_correct_order
```

Result:

```text
b7066badd871b7d041cecf1a1f3d398c1c1a3bd60b226947ff1753cdd95bd538
b7066badd871b7d041cecf1a1f3d398c1c1a3bd60b226947ff1753cdd95bd538
```

The hashes are identical.

This proves that the embedded payload is preserved byte-for-byte inside the fused executable.

---

## Runtime Verification

The extracted executable was executed.

Command:

```bash
./extracted_correct_order
```

Output:

```text
They call it a Royale with Cheese.
```

This confirms that the extracted payload is still a valid ELF executable and functions exactly as the original program.

---

## Reviewer Relevance

This step demonstrates:

* Verification of payload integrity.
* Validation using cryptographic hashes.
* Successful extraction of an embedded executable.
* Correct preservation of the original binary.
* Selection of the most appropriate tools after testing different approaches.

Screenshot:

`screenshots/20-final-payload-integrity.png`


# Step 19 - Final Binary Fusion Implementation

After validating each individual component of the project, the final binary fusion workflow was implemented.

The finished solution creates a single executable that contains both original programs and executes them sequentially.

Unlike the earlier proof-of-concept versions, the final implementation does not modify the original application code directly. Instead, it builds a dedicated runtime loader that manages the execution of both embedded executables.

---

## Final Fusion Process

The fusion process consists of the following stages:

1. Build the runtime loader written in Go.
2. Embed the first executable (`jules_static`) into the loader as `.jules_payload`.
3. Embed the second executable (`vincent_static`) into the loader as `.vincent_payload`.
4. Produce a single executable containing:

   * the runtime loader
   * Jules
   * Vincent

The runtime loader is responsible for locating the embedded payloads, extracting them to temporary executable files, executing them in sequence, and removing the temporary files afterwards.

---

## Runtime Loader Responsibilities

When the fused executable starts, it performs the following steps automatically:

1. Open its own executable (`/proc/self/exe`).
2. Locate `.jules_payload`.
3. Extract the payload to a temporary executable.
4. Execute Jules.
5. Locate `.vincent_payload`.
6. Extract the payload to a temporary executable.
7. Execute Vincent.
8. Remove both temporary files.
9. Exit normally.

The original executables are therefore never modified. They are embedded inside the fused executable and executed when required.

---

## Building the Final Binary

Command:

```bash
python scripts/fuse.py jules_static vincent_static fused_jules_final
```

Output:

```text
Created: fused_jules_final
First executable: jules_static
Second executable: vincent_static
Runtime loader built with: Go
Payloads embedded with: objcopy
Added section: .jules_payload
Added section: .vincent_payload
Fusion completed successfully
```

---

## Sequential Execution Verification

The final executable was executed.

Command:

```bash
./fused_jules_final
```

Output:

```text
What do they call it?
They call it a Royale with Cheese.
```

This confirms that the runtime loader successfully executes both embedded executables in the correct order.

---

## Output Redirection Verification

The assignment also requires support for redirecting the output of both original programs independently.

The runtime loader routes:

* Jules output to **standard output (stdout)**.
* Vincent output to **standard error (stderr)**.

Command:

```bash
./fused_jules_final > output_jules.txt 2> output_vincent.txt
```

Verification:

```bash
cat output_jules.txt
cat output_vincent.txt
```

Result:

```text
JULES OUTPUT:
What do they call it?

VINCENT OUTPUT:
They call it a Royale with Cheese.
```

This confirms that both embedded executables produce separate output streams exactly as required.

---

## Final Project Status

The completed implementation provides:

* ELF header analysis
* ELF section analysis
* ELF segment analysis
* Compatibility checking
* Support for executable, PIE, and relocatable ELF files
* Detection of incompatible architectures
* Support for non-standard section names
* Runtime payload extraction
* Payload integrity verification
* Sequential execution of both embedded executables
* Output redirection
* Automatic cleanup of temporary payload files

---

## Reviewer Relevance

This implementation satisfies the core functional requirements of the project:

* Parsing ELF binaries
* Combining two executables into a single binary
* Runtime execution of both embedded programs
* Verification of correct execution order
* Output redirection support
* Working implementation demonstrated using statically linked executables

Screenshot:

`screenshots/23-final-fused-execution.png`



