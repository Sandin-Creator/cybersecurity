import os
import shutil
import subprocess
import sys
import tempfile

import lief


def fail(message: str) -> None:
    print(f"Error: {message}")
    sys.exit(1)


if len(sys.argv) != 4:
    print(
        "Usage: python scripts/fuse.py "
        "<base_binary> <payload_binary> <output_binary>"
    )
    sys.exit(1)

base_path = os.path.abspath(sys.argv[1])
payload_path = os.path.abspath(sys.argv[2])
output_path = os.path.abspath(sys.argv[3])

if not os.path.isfile(base_path):
    fail(f"base binary does not exist: {base_path}")

if not os.path.isfile(payload_path):
    fail(f"payload binary does not exist: {payload_path}")

if shutil.which("objcopy") is None:
    fail("objcopy is not installed or is not available in PATH")

payload_size = os.path.getsize(payload_path)

if payload_size == 0:
    fail("payload binary is empty")

output_directory = os.path.dirname(output_path)

if not output_directory:
    output_directory = "."

temporary_file = None

try:
    # Create a temporary ELF file in the output directory.
    with tempfile.NamedTemporaryFile(
        prefix="binary_fusion_",
        suffix=".elf",
        dir=output_directory,
        delete=False,
    ) as temp:
        temporary_file = temp.name

    # objcopy preserves the complete payload byte-for-byte inside
    # the non-loaded .fused_payload resource section.
    objcopy_command = [
        "objcopy",
        "--add-section",
        f".fused_payload={payload_path}",
        "--set-section-flags",
        ".fused_payload=contents,readonly",
        base_path,
        temporary_file,
    ]

    objcopy_result = subprocess.run(
        objcopy_command,
        capture_output=True,
        text=True,
        check=False,
    )

    if objcopy_result.returncode != 0:
        fail(
            "objcopy failed:\n"
            + (objcopy_result.stderr or objcopy_result.stdout)
        )

    binary = lief.parse(temporary_file)

    if binary is None:
        fail("LIEF failed to parse the intermediate ELF")

    original_entry = binary.entrypoint

    # Temporary safe x86-64 Linux loader:
    #
    # mov eax, 60   ; sys_exit
    # xor edi, edi  ; exit status 0
    # syscall
    #
    # This will later be replaced by the real sequential loader.
    exit_stub = [
        0xB8,
        0x3C,
        0x00,
        0x00,
        0x00,
        0x31,
        0xFF,
        0x0F,
        0x05,
    ]

    loader_section = lief.ELF.Section(".fused_loader")
    loader_section.content = exit_stub
    loader_section.type = lief.ELF.Section.TYPE.PROGBITS
    loader_section.flags = (
        lief.ELF.Section.FLAGS.ALLOC
        | lief.ELF.Section.FLAGS.EXECINSTR
    )

    added_loader = binary.add(loader_section, loaded=True)

    if added_loader is None:
        fail("LIEF failed to add .fused_loader")

    loader_address = added_loader.virtual_address

    if loader_address == 0:
        fail(".fused_loader did not receive a valid virtual address")

    binary.header.entrypoint = loader_address

    binary.write(output_path)
    os.chmod(output_path, 0o755)

    print(f"Created: {output_path}")
    print(f"Base binary: {base_path}")
    print(f"Payload binary: {payload_path}")
    print(f"Original entry point: {hex(original_entry)}")
    print(f"Loader entry point: {hex(loader_address)}")
    print(f"Payload size: {payload_size} bytes")
    print("Payload embedded with: objcopy")
    print("Added section: .fused_payload")
    print("Added section: .fused_loader")
    print("Entry point modified: YES")
    print("Loader behavior: exit successfully with status 0")

finally:
    if temporary_file and os.path.exists(temporary_file):
        os.remove(temporary_file)