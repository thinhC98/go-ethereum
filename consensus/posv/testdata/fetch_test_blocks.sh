#!/bin/bash
# Script to fetch specific block ranges from Viction mainnet for testing

set -e

cd "$(dirname "$0")"

echo "=== Fetching Viction Mainnet Blocks for Testing ==="
echo ""

# Fetch epoch 1 (blocks 900-1799) - contains blocks 900-950
echo "📦 Fetching epoch 1 (blocks 900-1799)..."
go run dump_epoch.go -epoch 1 -output epoch_1.json

echo ""
echo "✓ Download complete!"
echo ""
echo "Files created:"
ls -lh epoch_*.json

echo ""
echo "=== Usage in Tests ==="
echo ""
echo "// Load all blocks from epoch 1 (900-1799)"
echo 'headers, _ := LoadEpochFromFile("testdata/epoch_1.json")'
echo ""
echo "// Load specific range (900-950)"
echo 'headers, _ := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)'
echo ""
echo "Now you can run: go test -v ./consensus/posv/... -run TestLoad"
