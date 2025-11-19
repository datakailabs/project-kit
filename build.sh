#!/usr/bin/env bash
# Build script for PK (Project Kit)

set -e  # Exit on error

echo "🔨 Building PK (Project Kit)..."
echo

# Navigate to project directory
cd "$(dirname "$0")"

# Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy

# Build binary
echo "🏗️  Compiling..."
go build -o bin/pk .

# Make executable
chmod +x bin/pk

echo
echo "✅ Build complete!"
echo "Binary location: $(pwd)/bin/pk"
echo
echo "To install globally:"
echo "  sudo mv bin/pk /usr/local/bin/pk"
echo
echo "To test:"
echo "  ./bin/pk --help"
echo "  ./bin/pk list"
