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
            })
    
    for section_name, symbols in section_symbols.items():
        for sym in symbols:
            if section_name != "" and sym['size'] > 0:
                print(section_name, sym['name'], sym['size'], sym['type'])

def print_macho_info(path):
    macho = lief.parse(path)

    # Build section map (Mach-O: section_number starts at 1, not 0! Index 0 = NO_SECT)
    section_map = {i + 1: s for i, s in enumerate(macho.sections)}  # Indexing from 1

    section_symbols = defaultdict(list)

    for section in macho.sections:
        print(section.segment.name, section.name, section.size, section.type.name if section.type else "N/A")
        
if __name__ == "__main__":
    for f in sys.argv[1:]:
        bin = lief.parse(f)
        if bin.format.name == "ELF":
            print_elf_info(f)
        elif bin.format.name == "MACHO":
            print_macho_info(f)
        else:
            print(f"{f}: Not a supported binary (ELF or Mach-O).")
