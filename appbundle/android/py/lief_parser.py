import sys
import lief

def print_elf_info(path):
    elf = lief.parse(path)
    print(f"\n--- ELF file: {path} ---")
    print("Sections:")
    for sec in elf.sections:
        print(sec.name, sec.type, sec.type.name if sec.type else "")
    print("Segments:")
    for seg in elf.segments:
        print(seg.type.name if seg.type else "", seg.file_size)

def print_macho_info(path):
    macho = lief.parse(path)
    print(f"\n--- Mach-O file: {path} ---")
    print("Sections:")
    for sec in macho.sections:
        print(f"  {sec.name:20s} offset={sec.offset} size={sec.size} segment={sec.segment.name}")
    print("Segments:")
    for seg in macho.segments:
        print(f"  {seg.name:20s} file_offset={seg.file_offset} file_size={seg.file_size} virtual_addr={hex(seg.virtual_address)}")

if __name__ == "__main__":
    for f in sys.argv[1:]:
        bin = lief.parse(f)
        if bin.format.name == "ELF":
            print_elf_info(f)
        elif bin.format.name == "MACHO":
            print_macho_info(f)
        else:
            print(f"{f}: Not a supported binary (ELF or Mach-O).")
