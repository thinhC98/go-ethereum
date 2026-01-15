# PoSV Blockchain Testing Tools

This directory contains tools for fetching and testing with Viction mainnet blocks.

## Tools Overview

### 1. `dump_epoch.go`
Fetches a complete epoch (900 blocks) from Viction mainnet.

**Usage:**
```bash
# Fetch epoch 1 (blocks 900-1799)
go run dump_epoch.go -epoch 1 -output epoch_1.json

# Fetch epoch 10 (blocks 9000-9899)
go run dump_epoch.go -epoch 10 -output epoch_10.json

# Use custom RPC endpoint
go run dump_epoch.go -epoch 1 -rpc https://rpc.viction.xyz -output epoch_1.json
```

### 2. `dump_snapshot.go`
Fetches blocks and analyzes snapshot/validator data.

### 3. `fetch_test_blocks.sh`
Quick script to fetch commonly used test data.

**Usage:**
```bash
cd testdata
./fetch_test_blocks.sh
```

## Using Block Data in Tests

### Method 1: Generate Test Blocks Programmatically

```go
func TestMyFeature(t *testing.T) {
    engine := New(nil, nil)
    
    // Create blocks 900-950
    db, genesis, chain, err := createTestChainFromRange(engine, 900, 950, "hash")
    if err != nil {
        t.Fatalf("Failed to create test chain: %v", err)
    }
    defer chain.Stop()
    
    // Now test with the blockchain
    // ...
}
```

### Method 2: Load Real Blocks from Viction Mainnet

```go
func TestWithRealData(t *testing.T) {
    // Load all blocks from epoch 1 (900-1799)
    headers, err := LoadEpochFromFile("testdata/epoch_1.json")
    if err != nil {
        t.Skip("Epoch file not found")
    }
    
    // Or load specific range (900-950)
    headers, err := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
    if err != nil {
        t.Skip("Epoch file not found")
    }
    
    // Process headers
    for _, header := range headers {
        fmt.Printf("Block: %s Hash: %s\n", header.Number, header.Hash)
    }
}
```

## Block Ranges Explained

In PoSV (Viction), blocks are organized in epochs of 900 blocks:

- **Epoch 0**: Blocks 0-899
- **Epoch 1**: Blocks 900-1799
- **Epoch 2**: Blocks 1800-2699
- **Epoch N**: Blocks N×900 to (N+1)×900-1

Checkpoint blocks occur at:
- Block 900 (end of epoch 0)
- Block 1800 (end of epoch 1)
- Block 2700 (end of epoch 2)
- etc.

## Example: Testing Blocks 900-950

### Step 1: Fetch the Data

```bash
cd consensus/posv/testdata
go run dump_epoch.go -epoch 1 -output epoch_1.json
```

This fetches blocks 900-1799 (entire epoch 1).

### Step 2: Use in Your Test

```go
func TestBlocks900to950(t *testing.T) {
    // Load the specific range you need
    headers, err := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
    if err != nil {
        t.Fatalf("Failed to load blocks: %v", err)
    }
    
    // You now have 51 block headers (900-950 inclusive)
    fmt.Printf("Loaded %d blocks\n", len(headers)) // Output: 51
    
    // Test your PoSV logic
    for i, header := range headers {
        // Convert to types.Header if needed
        // Test snapshot.apply(), validator checks, etc.
    }
}
```

## Available Helper Functions

### In `blockchain_test.go`:

```go
// Create test blockchain with range
createTestChainFromRange(engine, 900, 950, "hash")

// Generate blocks with range
makeBlockChainFromRange(genesis, 900, 950, engine, seed)

// Load all blocks from epoch file
LoadEpochFromFile("testdata/epoch_1.json")

// Load specific block range
LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
```

## File Format

The JSON files have this structure:

```json
{
  "startBlock": 900,
  "endBlock": 1799,
  "description": "Complete epoch 1 from Viction mainnet",
  "headers": [
    {
      "number": "0x384",
      "hash": "0x...",
      "parentHash": "0x...",
      "miner": "0x...",
      "extraData": "0x...",
      "difficulty": "0x2",
      "gasLimit": "0x...",
      "gasUsed": "0x...",
      "timestamp": "0x..."
    }
  ],
  "chainConfig": {
    "chainId": 88,
    "epoch": 900
  }
}
```

## Testing Patterns

### Test Epoch Transitions
```go
// Load blocks around checkpoint (e.g., 899-901)
headers, _ := LoadBlockRangeFromFile("testdata/epoch_1.json", 899, 901)
// Test snapshot behavior across epoch boundary
```

### Test Validator Changes
```go
// Load full epoch to see validator rotations
headers, _ := LoadEpochFromFile("testdata/epoch_1.json")
// Analyze validator changes throughout epoch
```

### Test Specific Block Range
```go
// Load exactly the blocks you need for testing
headers, _ := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
// Test with 51 blocks
```

## Running Tests

```bash
# Run all PoSV tests
go test -v ./consensus/posv/...

# Run specific test
go test -v ./consensus/posv/... -run TestLoadBlockRange

# Run with race detection
go test -race ./consensus/posv/...
```

## Notes

- Each epoch file is ~2-5 MB depending on block content
- Fetching 900 blocks takes ~2-5 minutes
- Files are reusable - fetch once, use in many tests
- Use `.gitignore` to avoid committing large JSON files
- For CI/CD, consider fetching on-demand or using fixtures
