# Quick Start: Testing with Blocks 900-950

This guide shows you exactly how to fetch and test with blocks 900-950 from Viction mainnet.

## Option 1: Use Real Mainnet Data

### Step 1: Fetch the blocks
```bash
cd consensus/posv/testdata
go run dump_epoch.go -epoch 1 -output epoch_1.json
```

This will download blocks 900-1799 (the entire epoch 1) which includes your target range 900-950.

### Step 2: Use in your test
```go
func TestMyFeature(t *testing.T) {
    // Load blocks 900-950 from the fetched data
    headers, err := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
    if err != nil {
        t.Skip("Epoch file not found, run: go run testdata/dump_epoch.go -epoch 1")
    }
    
    fmt.Printf("Loaded %d blocks from Viction mainnet\n", len(headers)) // 51 blocks
    
    // Now test with real block data
    for _, header := range headers {
        // Your test logic here
        fmt.Printf("Block %s: %s\n", header.Number, header.Hash)
    }
}
```

## Option 2: Generate Test Data Programmatically

This is faster and doesn't require network access:

```go
func TestWithGeneratedBlocks(t *testing.T) {
    engine := New(nil, nil)
    
    // Generate blocks 900-950
    db, genesis, chain, err := createTestChainFromRange(engine, 900, 950, "hash")
    if err != nil {
        t.Fatalf("Failed to create test chain: %v", err)
    }
    defer chain.Stop()
    
    // Test with the blockchain
    currentBlock := chain.CurrentBlock()
    fmt.Printf("Current block: %d\n", currentBlock.Number.Uint64()) // 51
    
    // Your test logic here
}
```

## Quick Commands

```bash
# Fetch blocks 900-950 (via epoch 1: 900-1799)
cd consensus/posv/testdata && go run dump_epoch.go -epoch 1 -output epoch_1.json

# Run your tests
go test -v ./consensus/posv/... -run TestLoadBlockRange

# Run all PoSV tests
go test ./consensus/posv/...
```

## Why Blocks 900-950?

- Block 900 is the first block of epoch 1
- Block 899 is the last block of epoch 0 (checkpoint block)
- This range is useful for testing epoch transitions and validator changes
- It's a manageable size (51 blocks) for testing

## Available Helper Functions

| Function | Purpose | Example |
|----------|---------|---------|
| `LoadEpochFromFile` | Load all blocks from an epoch | `LoadEpochFromFile("epoch_1.json")` |
| `LoadBlockRangeFromFile` | Load specific block range | `LoadBlockRangeFromFile("epoch_1.json", 900, 950)` |
| `createTestChainFromRange` | Generate test blockchain | `createTestChainFromRange(engine, 900, 950, "hash")` |
| `makeBlockChainFromRange` | Generate blocks only | `makeBlockChainFromRange(genesis, engine, 900, 950, seed)` |

## File Structure

```
consensus/posv/
├── testdata/
│   ├── dump_epoch.go           # Tool to fetch epoch data
│   ├── dump_snapshot.go        # Tool to analyze snapshots
│   ├── fetch_test_blocks.sh    # Quick fetch script
│   ├── README.md               # Detailed documentation
│   └── epoch_1.json           # Downloaded data (900-1799)
├── blockchain_test.go          # Helper functions
└── blockchain_range_test.go    # Example tests
```

## Next Steps

1. **Fetch the data**: `cd testdata && go run dump_epoch.go -epoch 1 -output epoch_1.json`
2. **Look at the examples**: Check [blockchain_range_test.go](blockchain_range_test.go)
3. **Write your tests**: Use the helper functions shown above
4. **Run tests**: `go test -v ./consensus/posv/...`

For more details, see [testdata/README.md](testdata/README.md)
