# Testing with Block Ranges - Summary

## What Was Created

### 1. Helper Functions (blockchain_test.go)
- `LoadEpochFromFile(filename)` - Load all blocks from epoch JSON file
- `LoadBlockRangeFromFile(filename, start, end)` - Load specific block range

### 2. Test Examples (blockchain_range_test.go)
- `TestLoadBlocksFromEpochFile` - Shows how to load epoch data
- `TestLoadBlockRangeFromFile` - Shows how to load specific ranges (900-950)
- `TestUsageExamples` - Complete usage examples

### 3. Data Fetching Tools (testdata/)
- `dump_epoch.go` - Fetch complete epochs from Viction mainnet
- `fetch_test_blocks.sh` - Quick script to fetch test data
- `README.md` - Complete documentation

## How to Use for Testing Blocks 900-950

### Step 1: Fetch the Data
```bash
cd consensus/posv/testdata
go run dump_epoch.go -epoch 1 -output epoch_1.json
```

This fetches blocks 900-1799 from Viction mainnet.

### Step 2: Load in Your Tests
```go
// Load blocks 900-950 from the epoch file
headers, err := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
if err != nil {
    t.Fatalf("Failed to load blocks: %v", err)
}

// Now you have 51 block headers (900-950 inclusive)
for _, header := range headers {
    var blockNum uint64
    fmt.Sscanf(header.Number, "0x%x", &blockNum)
    
    // Use the header data for testing
    fmt.Printf("Block %d: %s\n", blockNum, header.Hash)
}
```

## Current Limitations

1. **createTestChainFromRange** is currently disabled due to `consensus.Engine` interface compatibility issues between Posv and the test framework
2. The Posv engine doesn't fully implement `consensus.Engine` (missing Close method, wrong ChainReader type)
3. For now, focus on using `LoadBlockRangeFromFile` to get real block data for testing

## Working Tests

Run these to verify everything works:
```bash
# Test loading epoch files (will skip if file doesn't exist)
go test -v ./consensus/posv -run TestLoadBlocksFromEpochFile

# Test loading specific range
go test -v ./consensus/posv -run TestLoadBlockRange

# See usage examples
go test -v ./consensus/posv -run TestUsageExamples
```

## Next Steps

To make `createTestChainFromRange` work, you would need to:
1. Make Posv implement `consensus.Engine` interface properly
2. Update ChainReader types to match `consensus.ChainHeaderReader`
3. Add the `Close()` method to Posv

But for now, **the JSON loading functions work perfectly** and you can use them to test with real Viction mainnet data!
