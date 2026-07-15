import sys
import os
import lief

if len(sys.argv) != 4:
    print("Usage: python scripts/fuse.py <base_binary> <payload_binary> <output_binary>")
    sys.exit(1)

base_path = sys.argv[1]
payload_path = sys.argv[2]
output_path = sys.argv[3]

base = lief.parse(base_path)

if base is None:
    print("Error: failed to parse base binary")
    sys.exit(1)

original_entry = base.entrypoint

with open(payload_path, "rb") as f:
    payload_data = list(f.read())

payload_section = lief.ELF.Section(".fused_payload")
payload_section.content = payload_data
payload_section.type = lief.ELF.Section.TYPE.PROGBITS
payload_section.flags = lief.ELF.Section.FLAGS.ALLOC

base.add(payload_section, loaded=False)

loader_section = lief.ELF.Section(".fused_loader")
loader_section.content = [0x90] * 128
loader_section.type = lief.ELF.Section.TYPE.PROGBITS
loader_section.flags = (
    lief.ELF.Section.FLAGS.ALLOC |
    lief.ELF.Section.FLAGS.EXECINSTR
)

base.add(loader_section, loaded=True)

base.write(output_path)
os.chmod(output_path, 0o755)

print(f"Created: {output_path}")
print(f"Base binary: {base_path}")
print(f"Payload binary: {payload_path}")
print(f"Original entry point: {hex(original_entry)}")
print(f"Payload size: {len(payload_data)} bytes")
print("Added section: .fused_payload")
print("Added section: .fused_loader")
print("Entry point modified: NO")