#!/usr/bin/env python3
# filepath: arsc_parser.py

import os
import sys
import csv
from androguard.core.resources import arsc
from androguard.core.resources.arsc import ARSCParser

def parse_arsc_file(file_path, output_format='csv'):
    """
    Parse an Android binary resource file (.arsc) and extract resource values.
    
    Args:
        file_path: Path to the .arsc file
        output_format: Output format (csv or plain)
    """
    if not os.path.exists(file_path):
        print(f"Error: File '{file_path}' not found.", file=sys.stderr)
        return False
    
    try:
        # Parse the ARSC file
        with open(file_path, 'rb') as f:
            arsc_data = f.read()
        arsc_parser = ARSCParser(arsc_data)
        
        # Process and output the resource values
        resources = []
        
        for package in arsc_parser.packages.values():
            package_name = package.name
            
            for type_config in package.types.values():
                if not type_config:
                    continue
                
                type_name = type_config[0].type
                
                for config in type_config:
                    for entry_key, entry in config.entries.items():
                        if not entry:
                            continue
                        
                        resource_name = package.get_string(entry.key_index)
                        
                        if entry.value:
                            if hasattr(entry.value, 'value'):
                                resource_value = entry.value.value
                                resources.append((package_name, type_name, resource_name, resource_value))
        
        # Output the collected resources
        if output_format == 'csv':
            writer = csv.writer(sys.stdout)
            writer.writerow(['Package', 'Type', 'Name', 'Value'])
            for resource in resources:
                writer.writerow(resource)
        else:
            for package_name, type_name, resource_name, resource_value in resources:
                print(f"Package: {package_name}")
                print(f"  Type: {type_name}")
                print(f"    Resource: {resource_name}")
                print(f"      Value: {resource_value}")
        
        return True
    except Exception as e:
        print(f"Error parsing ARSC file: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        return False

def main():
    if len(sys.argv) < 2:
        print("Usage: python arsc_parser.py <arsc_file> [csv|plain]")
        return 1
    
    arsc_file = sys.argv[1]
    output_format = 'csv'
    if len(sys.argv) > 2 and sys.argv[2].lower() == 'plain':
        output_format = 'plain'
    
    if parse_arsc_file(arsc_file, output_format):
        return 0
    else:
        return 1

if __name__ == "__main__":
    sys.exit(main())