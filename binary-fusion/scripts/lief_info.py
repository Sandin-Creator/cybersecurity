import sys
import lief

if len(sys.argv) != 2:
    print("Usage: python scripts/lief_info.py <binary>")
    sys.exit(1)

path = sys.argv[1]
binary = lief.parse(path)

if binary is None:
    print("Error: failed to parse binary")
    sys.exit(1)

print("File:", path)
print("Format:", binary.format)
print("Entry point:", hex(binary.entrypoint))
print("Architecture:", binary.header.machine_type)
print("Sections:", len(binary.sections))
print("Segments:", len(binary.segments))