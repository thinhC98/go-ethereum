# Snapshot Verification in PoSV Tests

This document explains how to use snapshot creation and verification when testing PoSV blockchain functionality.

## Overview

When creating a new blockchain for testing, you can now automatically retrieve and verify snapshots at specific block numbers. This helps ensure that the PoSV consensus state (signers, votes, etc.) is correctly maintained.

## Available Functions

### 1. `createPosvChainWithSnapshot()`

Creates a PoSV blockchain and automatically retrieves snapshots at specified checkpoints.

```go
func createPosvChainWithSnapshot(
    startBlock, endBlock int, 
    checkpoints []uint64
) (*core.BlockChain, *Posv, map[uint64]*Snapshot, error)
```

**Note:** Currently disabled due to interface compatibility issues. The PoSV engine doesn't fully implement `consensus.Engine` (missing Close method, wrong ChainReader type).

### 2. `getSnapshotFromExistingChain()`

Retrieves a snapshot from an existing blockchain at a specific block number.

```go
func getSnapshotFromExistingChain(
    chain *core.BlockChain, 
    engine *Posv, 
    blockNum uint64
) (*Snapshot, error)
```

**Usage:**
```go
// Create a blockchain
db, _, chain, err := createTestChainFromRange(engine, 900, 950, "hash")
if err != nil {
    // Handle error
}
defer chain.Stop()

// Get snapshot at block 925
snap, err := getSnapshotFromExistingChain(chain, engine, 925)
if err != nil {
    // Handle error
}

// Verify snapshot has signers
signers := snap.GetSigners()
fmt.Printf("Snapshot at block 925 has %d signers\n", len(signers))
```

### 3. `verifySnapshot()`

Retrieves and performs basic verification on a snapshot.

```go
func verifySnapshot(
    blockchain *core.BlockChain, 
    engine *Posv, 
    blockNum uint64
) (*Snapshot, error)
```

Checks:
- Header exists at the specified block number
- Snapshot can be retrieved
- Snapshot has at least one signer

### 4. `verifySnapshotData()`

Performs detailed verification of snapshot data against expected values.

```go
func verifySnapshotData(
    snap *Snapshot, 
    expectedSigners []common.Address
) error
```

Verifies:
- Snapshot is not nil
- Number of signers matches expectations
- All expected signers are present in the snapshot

## Testing Patterns

### Pattern 1: Verify Snapshot at Specific Blocks

```go
func TestSnapshotAtCheckpoint(t *testing.T) {
    config := &PosvConfig{
        Epoch:  900,
        Period: 2,
        Gap:    450,
    }
    engine := New(config, nil)
    
    // Create blockchain
    _, _, chain, err := createTestChainFromRange(engine, 0, 1000, "hash")
    if err != nil {
        t.Fatalf("Failed to create chain: %v", err)
    }
    defer chain.Stop()
    
    // Verify snapshot at epoch boundary (block 900)
    snap, err := verifySnapshot(chain, engine, 900)
    if err != nil {
        t.Errorf("Snapshot verification failed: %v", err)
    }
    
    signers := snap.GetSigners()
    t.Logf("Block 900 has %d signers", len(signers))
}
```

### Pattern 2: Compare Snapshots Across Blocks

```go
func TestSnapshotEvolution(t *testing.T) {
    config := &PosvConfig{
        Epoch:  900,
        Period: 2,
        Gap:    450,
    }
    engine := New(config, nil)
    
    _, _, chain, err := createTestChainFromRange(engine, 0, 1000, "hash")
    if err != nil {
        t.Fatalf("Failed to create chain: %v", err)
    }
    defer chain.Stop()
    
    // Get snapshots at multiple points
    checkpoints := []uint64{0, 450, 900}
    snapshots := make(map[uint64]*Snapshot)
    
    for _, blockNum := range checkpoints {
        snap, err := getSnapshotFromExistingChain(chain, engine, blockNum)
        if err != nil {
            t.Errorf("Failed to get snapshot at block %d: %v", blockNum, err)
            continue
        }
        snapshots[blockNum] = snap
    }
    
    // Compare signer sets across checkpoints
    for blockNum, snap := range snapshots {
        signers := snap.GetSigners()
        t.Logf("Block %d: %d signers", blockNum, len(signers))
    }
}
```

### Pattern 3: Verify with Mainnet Data

```go
func TestMainnetSnapshotData(t *testing.T) {
    // Load real block data
    headers, err := LoadBlockRangeFromFile("testdata/epoch_1.json", 900, 950)
    if err != nil {
        t.Skipf("Mainnet data not available: %v", err)
    }
    
    // For each block, you can extract expected signer information
    // and verify against snapshot data
    // (Implementation depends on your JSON structure)
}
```

## Current Limitations

### Interface Compatibility

The PoSV engine doesn't fully implement `consensus.Engine`:

1. **Missing Close() method** - Required by consensus.Engine interface
2. **ChainReader type mismatch** - PoSV uses local ChainReader, consensus package expects its own

**Workaround:** Tests currently use `ethash.NewFakeFailer()` as the consensus engine for blockchain creation, then use PoSV engine methods separately for snapshot operations.

### State Pruning

Creating blockchains with many blocks may fail with "pruned ancestor" errors due to state pruning. For testing:

- Keep test chains small (< 100 blocks) when possible
- Use mainnet data from JSON files for larger ranges
- Focus tests on specific block ranges rather than full epochs

## Recommended Testing Workflow

1. **Use Real Data for Integration Tests**
   ```bash
   cd testdata
   go run dump_epoch.go -epoch 1 -output epoch_1.json
   go run dump_snapshot.go -block 900 -output snapshot_900.json
   ```

2. **Use Generated Data for Unit Tests**
   ```go
   // Small, fast tests with generated blocks
   _, _, chain, _ := createTestChainFromRange(engine, 0, 50, "hash")
   ```

3. **Combine Both for Comprehensive Testing**
   ```go
   // Load expected data from mainnet
   headers := LoadBlockRangeFromFile("epoch_1.json", 900, 950)
   
   // Verify against generated blockchain
   snap := getSnapshotFromExistingChain(chain, engine, 900)
   // Compare snap data with headers data
   ```

## Examples

See these test files for complete examples:

- `blockchain_range_test.go` - Examples of all snapshot verification patterns
- `blockchain_test.go` - Helper function implementations
- `example_test.go` - End-to-end usage demonstrations

## Future Improvements

To enable full PoSV blockchain creation in tests:

1. Add `Close()` method to PoSV struct
2. Create adapter to make PoSV compatible with `consensus.Engine`
3. Or create test-specific PoSV wrapper that implements full interface

For now, the hybrid approach (ethash blockchain + PoSV snapshot methods) provides sufficient testing capabilities.
