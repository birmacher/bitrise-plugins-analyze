import sys
from androguard.misc import AnalyzeDex

if len(sys.argv) != 2:
    print("Usage: python dex_size.py [path_to_dex_file]")
    sys.exit(1)

try:
    hash, d, dx = AnalyzeDex(sys.argv[1])
    
    classes = d.get_classes()
    for cls in classes:
        class_name = cls.get_name()
        class_size = 0
        for m in cls.get_methods():
            method_name = m.get_name()
            if m.get_code() is not None:
                class_size += len(m.get_code().get_raw())
        print(class_name, class_size)
        
except Exception as e:
    print(f"Error processing file: {e}")
    sys.exit(1)
