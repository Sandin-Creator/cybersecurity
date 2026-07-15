# Binary Fusion

## Project Overview

This project implements a binary fusion experiment for ELF executables on Linux. It combines two statically linked executables into a single self-contained binary that extracts and executes both programs sequentially at runtime.

The implementation uses:

- Two sample executables: `jules_static` and `vincent_static`
- A Go-based ELF inspection tool located in `cmd/fuser`
- A runtime loader implemented in `loader/runtime_loader.go`
- A Python fusion script located in `scripts/fuse.py`
- LIEF for ELF parsing and modification experiments
- GNU `objcopy` for embedding executable payloads into ELF sections

## Features

- ELF header analysis
- ELF section analysis
- ELF segment analysis
- ELF compatibility checking
- Runtime extraction of embedded executables
- Sequential execution of two embedded executables
- Separate stdout and stderr output streams
- Payload integrity verification using SHA-256

## Requirements

- Go 1.22 or newer
- Python 3.10 or newer
- GCC
- GNU Binutils (`objcopy`)
- LIEF Python library

Install the Python dependency:

```bash
python -m venv venv
source venv/bin/activate
pip install lief
```

## Build Instructions

Compile the sample executables:

```bash
gcc -no-pie -static -o jules_static jules.c
gcc -no-pie -static -o vincent_static vincent.c
```

Build the Go tools:

```bash
go mod tidy

go build -o fuser_tool ./cmd/fuser
go build -o runtime_loader ./loader/runtime_loader.go
```

## Usage

### Analyze an ELF binary

```bash
go run ./cmd/fuser jules_static
```

### List ELF sections

```bash
go run ./cmd/fuser --sections jules_static
```

### List ELF segments

```bash
go run ./cmd/fuser --segments jules_static
```

### Create the fused executable

```bash
python scripts/fuse.py jules_static vincent_static fused_jules_final
```

### Run the fused executable

```bash
./fused_jules_final
```

Expected output:

```text
What do they call it?
They call it a Royale with Cheese.
```

### Redirect program output

```bash
./fused_jules_final > output_jules.txt 2> output_vincent.txt
```

Verify the outputs:

```bash
cat output_jules.txt
cat output_vincent.txt
```

Expected result:

```text
JULES OUTPUT:
What do they call it?

VINCENT OUTPUT:
They call it a Royale with Cheese.
```

## Notes

The runtime loader opens its own executable (`/proc/self/exe`), extracts both embedded executables into temporary files, executes them sequentially, removes the temporary files, and exits normally. The original executables remain embedded and unchanged inside the fused ELF binary.