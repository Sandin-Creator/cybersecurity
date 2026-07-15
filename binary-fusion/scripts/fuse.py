import os
import shutil
import subprocess
import sys
import tempfile


def fail(message: str) -> None:
    print(f"Error: {message}")
    sys.exit(1)


def run_command(command: list[str]) -> None:
    result = subprocess.run(
        command,
        capture_output=True,
        text=True,
        check=False,
    )

    if result.returncode != 0:
        fail(result.stderr or result.stdout or "command failed")


if len(sys.argv) != 4:
    print(
        "Usage: python scripts/fuse.py "
        "<first_binary> <second_binary> <output_binary>"
    )
    sys.exit(1)

first_path = os.path.abspath(sys.argv[1])
second_path = os.path.abspath(sys.argv[2])
output_path = os.path.abspath(sys.argv[3])

project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
runtime_source = os.path.join(
    project_root,
    "loader",
    "runtime_loader.go",
)

if not os.path.isfile(first_path):
    fail(f"first binary does not exist: {first_path}")

if not os.path.isfile(second_path):
    fail(f"second binary does not exist: {second_path}")

if not os.path.isfile(runtime_source):
    fail(f"runtime loader source does not exist: {runtime_source}")

if shutil.which("go") is None:
    fail("Go is not installed or is not available in PATH")

if shutil.which("objcopy") is None:
    fail("objcopy is not installed or is not available in PATH")

output_directory = os.path.dirname(output_path) or "."
temporary_files: list[str] = []

try:
    with tempfile.NamedTemporaryFile(
        prefix="binary_fusion_runtime_",
        dir=output_directory,
        delete=False,
    ) as temp:
        runtime_binary = temp.name
        temporary_files.append(runtime_binary)

    with tempfile.NamedTemporaryFile(
        prefix="binary_fusion_first_",
        dir=output_directory,
        delete=False,
    ) as temp:
        first_embedded_binary = temp.name
        temporary_files.append(first_embedded_binary)

    # Build the working self-contained runtime loader.
    run_command([
        "go",
        "build",
        "-o",
        runtime_binary,
        runtime_source,
    ])

    os.chmod(runtime_binary, 0o755)

    # Embed the first executable.
    run_command([
        "objcopy",
        "--add-section",
        f".jules_payload={first_path}",
        "--set-section-flags",
        ".jules_payload=contents,readonly",
        runtime_binary,
        first_embedded_binary,
    ])

    # Embed the second executable and create the final binary.
    run_command([
        "objcopy",
        "--add-section",
        f".vincent_payload={second_path}",
        "--set-section-flags",
        ".vincent_payload=contents,readonly",
        first_embedded_binary,
        output_path,
    ])

    os.chmod(output_path, 0o755)

    print(f"Created: {output_path}")
    print(f"First executable: {first_path}")
    print(f"Second executable: {second_path}")
    print("Runtime loader built with: Go")
    print("Payloads embedded with: objcopy")
    print("Added section: .jules_payload")
    print("Added section: .vincent_payload")
    print("Fusion completed successfully")

finally:
    for temporary_file in temporary_files:
        if os.path.exists(temporary_file):
            os.remove(temporary_file)