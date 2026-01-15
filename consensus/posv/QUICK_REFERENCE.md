# Quick Reference: Testing Blocks 900-950

## One-Line Commands

```bash
# Fetch mainnet data (run once)
cd consensus/posv/testdata && go run dump_epoch.go -epoch 1 -output epoch_1.json

# Run tests
go test -v ./consensus/posv/... -run TestBlocks900to950
```

## Code Snippets

### Load Real Mainnet Blocks (900-950)
```go
headers, err := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
// Returns 51 blocks from Viction mainnet
```

### Generate Test Blocks (900-950)
```go
engine := New(nil, nil)
_, _, chain, err := createTestChainFromRange(engine, 900, 950, "hash")
defer chain.Stop()
// Creates a test blockchain with 51 blocks
```

### Test Epoch Transition (899-901)
```go
headers, err := LoadBlockRangeFromFile("testdata/epoch_1.json", 899, 901)
// Block 899: checkpoint (end of epoch 0)
// Block 900: start of epoch 1
// Block 901: regular block
```

## File Structure
```
consensus/posv/
├── blockchain_test.go              # Helper functions
├── blockchain_range_test.go        # Example tests
├── example_test.go                 # Practical examples
├── TESTING_BLOCKS_900_950.md      # Quick start
├── SUMMARY.md                      # This implementation
├── QUICK_REFERENCE.md             # Cheat sheet
└── testdata/
    ├── dump_epoch.go               # Fetch tool
    ├── dump_snapshot.go            # Snapshot tool
    ├── fetch_test_blocks.sh        # Quick script
    ├── README.md                   # Detailed docs
    └── epoch_1.json               # Data file (900-1799)
```

## Helper Functions Cheat Sheet

| Function | Returns | Use Case |
|----------|---------|----------|
| `LoadEpochFromFile(file)` | All blocks in epoch | Full epoch testing |
| `LoadBlockRangeFromFile(file, start, end)` | Blocks in range | Specific range |
| `createTestChainFromRange(e, s, e, scheme)` | Blockchain | Generate test data |
| `makeBlockChainFromRange(g, e, s, e, seed)` | Blocks only | Custom chain |

## Block Numbers Quick Reference

| Block | Significance |
|-------|--------------|
| 899 | Last block of epoch 0 (checkpoint) |
| 900 | First block of epoch 1 |
| 950 | 51st block of epoch 1 |
| 1799 | Last block of epoch 1 (checkpoint) |
| 1800 | First block of epoch 2 |

## Common Test Patterns

### Pattern 1: Validate Block Range
```go
func TestValidateRange(t *testing.T) {
    headers, _ := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
    for _, h := range headers {
        var num uint64
        fmt.Sscanf(h.Number, "0x%x", &num)
        if num < 900 || num > 950 {
            t.Errorf("Block %d out of range", num)
        }
    }
}
```

### Pattern 2: Test Consensus Logic
```go
func TestConsensus(t *testing.T) {
    engine := New(nil, nil)
    _, _, chain, _ := createTestChainFromRange(engine, 900, 950, "hash")
    defer chain.Stop()
    
    current := chain.CurrentBlock()
    err := engine.VerifyHeader(chain, current, false)
    // Test verification logic
}
```

### Pattern 3: Analyze Validator Changes
```go
func TestValidators(t *testing.T) {
    headers, _ := LoadEpochFromFile("testdata/epoch_1.json")
    
    // Extract validator info from each block
    for _, h := range headers {
        if h.Validators != "" {
            // Parse and analyze validator set
        }
    }
}
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| File not found | Run: `cd testdata && go run dump_epoch.go -epoch 1` |
| Slow fetch | Normal, fetching 900 blocks takes 2-5 minutes |
| Test fails | Use `t.Skip()` if epoch file doesn't exist |
| Wrong block count | Check range: 900-950 inclusive = 51 blocks |

## Quick Checks

```bash
# Verify file exists
ls -lh consensus/posv/testdata/epoch_1.json

# Check JSON structure
head -50 consensus/posv/testdata/epoch_1.json

# Compile tests
go test -c ./consensus/posv/...

# Run specific test
go test -v ./consensus/posv/... -run TestLoadBlockRange
```
