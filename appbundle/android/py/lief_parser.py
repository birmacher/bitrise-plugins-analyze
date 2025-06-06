import sys
import lief
from collections import defaultdict

def print_elf_info(path):
    elf = lief.parse(path)

    section_map = {i: s for i, s in enumerate(elf.sections)}

    # Dict: section name -> list of symbols
    section_symbols = defaultdict(list)

    for sym in elf.symbols:
        # Section number/index for the symbol
        idx = sym.shndx
        # Only valid section indices (and non-empty symbol names)
        if idx >= 0 and idx in section_map and sym.name:
            section = section_map[idx]
            section_symbols[section.name].append({
                "name": sym.name,
                "size": sym.size,
                "type": sym.type.name if sym.type else "",
                "binding": sym.binding.name if sym.binding else ""
            })
    
    for section_name, symbols in section_symbols.items():
        for sym in symbols:
            if section_name != "" and sym['size'] > 0:
                print(section_name, sym['name'], sym['size'], sym['type'], sym['binding'])

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
