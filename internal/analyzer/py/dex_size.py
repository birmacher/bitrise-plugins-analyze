import sys
from androguard.core.analysis.analysis import Analysis
from androguard.misc import AnalyzeAPK, AnalyzeDex

if len(sys.argv) != 2:
    print("Usage: python dex_size.py [path_to_dex_file]")
    sys.exit(1)

try:
    # Try to analyze directly as a DEX file
    # Try to analyze directly as a DEX file
    hash, d, dx = AnalyzeDex(sys.argv[1])
    print("Dex file loaded successfully.")
    
    # Get classes and their sizes
    classes = d.get_classes()
    print("Number of classes:", len(classes))
    
    # Print class names and sizes
    for cls in classes:
        class_name = cls.get_name()
        # Get method information for size estimation
        methods = cls.get_methods()
        class_size = sum(m.get_length() for m in methods)
        print(class_name, class_size)
        
except Exception as e:
    print(f"Error processing file: {e}")
    sys.exit(1)