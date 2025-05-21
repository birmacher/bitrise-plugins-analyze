#!/bin/bash
set -e

# Set working directory to project root
cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Create bin directory if it doesn't exist
mkdir -p bin

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod tidy

# Build the application
echo "Building application..."
go build -o bin/ .

echo "✅ Setup complete! Binary has been built to the bin/ directory."

# Run the application
echo "Running application..."

bin/bitrise-plugins-analyze analyze ./scripts/HexaCalc.ipa