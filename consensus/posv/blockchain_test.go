package posv

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

// So we can deterministically seed different blockchains
var (
	canonicalSeed = 1
	forkSeed      = 2
)

func newCanonical(engine consensus.Engine, n int, full bool, scheme string) (ethdb.Database, *core.Genesis, *core.BlockChain, error) {
	var (
		genesis = &core.Genesis{
			BaseFee: big.NewInt(params.InitialBaseFee),
			Config: &params.ChainConfig{
				ChainID:             big.NewInt(1),
				HomesteadBlock:      big.NewInt(0),
				EIP150Block:         big.NewInt(0),
				EIP155Block:         big.NewInt(0),
				EIP158Block:         big.NewInt(0),
				ByzantiumBlock:      big.NewInt(0),
				ConstantinopleBlock: big.NewInt(0),
				PetersburgBlock:     big.NewInt(0),
				IstanbulBlock:       big.NewInt(0),
				BerlinBlock:         big.NewInt(0),
				// London block (EIP1559) is NOT set, avoiding BaseFee calculation issues
				LondonBlock: nil,
				Ethash:      new(params.EthashConfig),
			},
		}
	)

	cacheConfig := &core.CacheConfig{
		TrieCleanLimit:    256,
		TrieDirtyLimit:    256,
		TrieDirtyDisabled: true, // Disables pruning (archive mode)
		TrieTimeLimit:     5 * time.Minute,
		SnapshotLimit:     256,
		StateScheme:       scheme,
	}

	// Create and inject the requested chain
	if n == 0 {
		blockchain, _ := core.NewBlockChain(rawdb.NewMemoryDatabase(), cacheConfig, genesis, nil, engine, vm.Config{}, nil)
		return rawdb.NewMemoryDatabase(), genesis, blockchain, nil
	}

	if full {
		// Full block-chain requested
		// Create blockchain first with empty database
		db := rawdb.NewMemoryDatabase()
		blockchain, _ := core.NewBlockChain(db, core.DefaultCacheConfigWithScheme(scheme), genesis, nil, engine, vm.Config{}, nil)
		// Generate blocks using the blockchain's database
		blocks, _ := core.GenerateChain(genesis.Config, genesis.ToBlock(), engine, db, n, func(i int, b *core.BlockGen) {
			b.SetCoinbase(common.Address{0: byte(canonicalSeed), 19: byte(i)})
		})
		// Insert the blocks
		_, err := blockchain.InsertChain(blocks)
		return db, genesis, blockchain, err
	}
	// Header-only chain requested
	genDb, headers := makeHeaderChainWithGenesis(genesis, n, engine, canonicalSeed)
	blockchain, _ := core.NewBlockChain(genDb, cacheConfig, genesis, nil, engine, vm.Config{}, nil)
	_, err := blockchain.InsertHeaderChain(headers)
	return genDb, genesis, blockchain, err
}

// makeBlockChainWithGenesis creates a deterministic chain of blocks from genesis
func makeBlockChainWithGenesis(genesis *core.Genesis, n int, engine consensus.Engine, seed int) (ethdb.Database, []*types.Block) {
	db, blocks, _ := core.GenerateChainWithGenesis(genesis, engine, n, func(i int, b *core.BlockGen) {
		b.SetCoinbase(common.Address{0: byte(seed), 19: byte(i)})
	})
	return db, blocks
}

// makeHeaderChainWithGenesis creates a deterministic chain of headers from genesis.
func makeHeaderChainWithGenesis(genesis *core.Genesis, n int, engine consensus.Engine, seed int) (ethdb.Database, []*types.Header) {
	db, blocks := makeBlockChainWithGenesis(genesis, n, engine, seed)
	headers := make([]*types.Header, len(blocks))
	for i, block := range blocks {
		headers[i] = block.Header()
	}
	return db, headers
}

// makeBlockChainFromRange creates a blockchain with blocks from startBlock to endBlock
func makeBlockChainFromRange(genesis *core.Genesis, engine consensus.Engine, startBlock, endBlock int, seed int) (ethdb.Database, []*types.Block) {
	numBlocks := endBlock - startBlock + 1
	db, blocks, _ := core.GenerateChainWithGenesis(genesis, engine, numBlocks, func(i int, b *core.BlockGen) {
		// i starts from 0, so actual block number would be startBlock + i
		blockNum := startBlock + i
		b.SetCoinbase(common.Address{0: byte(seed), 19: byte(blockNum)})
	})
	return db, blocks
}

// createTestChainFromRange creates a full blockchain with blocks from 0 to endBlock
// Example: createTestChainFromRange(nil, 900, 950, "hash") creates blocks 0-950
// You can then access blocks in the range [startBlock, endBlock] from the blockchain
func createTestChainFromRange(_engine *Posv, startBlock, endBlock int, scheme string) (ethdb.Database, *core.Genesis, *core.BlockChain, error) {
	testEngine := ethash.NewFaker()

	// Use the newCanonical pattern for creating blockchain
	return newCanonical(testEngine, endBlock+1, true, scheme)
}

// createPosvChainWithSnapshot creates a PoSV blockchain and verifies snapshot at specific blocks
// Returns the blockchain, PoSV engine, and snapshot for verification
func createPosvChainWithSnapshot(startBlock, endBlock int, checkpoints []uint64) (*core.BlockChain, *Posv, map[uint64]*Snapshot, error) {
	// Note: This function is currently disabled due to consensus.Engine interface compatibility
	// The Posv engine doesn't implement consensus.Engine (missing Close method, wrong ChainReader type)
	// To enable this, we need to use a wrapper or modify the Posv struct (but user requested no consensus file changes)
	return nil, nil, nil, fmt.Errorf("PoSV blockchain creation disabled - interface compatibility issues")
}

// getSnapshotFromExistingChain retrieves and verifies a snapshot from an existing blockchain
// This is a practical workaround to test snapshot functionality without PoSV blockchain creation
func getSnapshotFromExistingChain(chain *core.BlockChain, engine *Posv, blockNum uint64) (*Snapshot, error) {
	header := chain.GetHeaderByNumber(blockNum)
	if header == nil {
		return nil, fmt.Errorf("header not found at block %d", blockNum)
	}

	// Get snapshot using PoSV engine
	snap, err := engine.snapshot(chain, blockNum, header.Hash(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot at block %d: %w", blockNum, err)
	}

	return snap, nil
}

// verifySnapshot checks snapshot data at a specific block
func verifySnapshot(blockchain *core.BlockChain, engine *Posv, blockNum uint64) (*Snapshot, error) {
	header := blockchain.GetHeaderByNumber(blockNum)
	if header == nil {
		return nil, fmt.Errorf("header not found at block %d", blockNum)
	}

	snap, err := engine.snapshot(blockchain, blockNum, header.Hash(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Verify snapshot has signers
	signers := snap.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("snapshot has no signers at block %d", blockNum)
	}

	return snap, nil
}

// verifySnapshotData performs detailed verification of snapshot
func verifySnapshotData(snap *Snapshot, expectedSigners []common.Address) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	signers := snap.GetSigners()
	if len(signers) != len(expectedSigners) {
		return fmt.Errorf("expected %d signers, got %d", len(expectedSigners), len(signers))
	}

	// Verify each expected signer is in the snapshot
	for _, expected := range expectedSigners {
		found := false
		for _, signer := range signers {
			if signer == expected {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected signer %s not found in snapshot", expected.Hex())
		}
	}

	return nil
}

// BlockHeaderJSON represents a block header from JSON dump
type BlockHeaderJSON struct {
	Number      string `json:"number"`
	Hash        string `json:"hash"`
	ParentHash  string `json:"parentHash"`
	Coinbase    string `json:"miner"`
	Extra       string `json:"extraData"`
	Validator   string `json:"validator,omitempty"`
	Penalties   string `json:"penalties,omitempty"`
	Validators  string `json:"validators,omitempty"`
	Difficulty  string `json:"difficulty"`
	GasLimit    string `json:"gasLimit"`
	GasUsed     string `json:"gasUsed"`
	Time        string `json:"timestamp"`
	Nonce       string `json:"nonce"`
	Root        string `json:"stateRoot"`
	TxHash      string `json:"transactionsRoot"`
	ReceiptHash string `json:"receiptsRoot"`
}

// EpochDataJSON represents the JSON structure from dump_epoch.go
type EpochDataJSON struct {
	StartBlock  uint64            `json:"startBlock"`
	EndBlock    uint64            `json:"endBlock"`
	Description string            `json:"description"`
	Headers     []BlockHeaderJSON `json:"headers"`
}

// LoadEpochFromFile loads epoch data from a JSON file (created by dump_epoch.go)
// Returns the headers that can be used for testing
func LoadEpochFromFile(filename string) ([]BlockHeaderJSON, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var epochData EpochDataJSON
	if err := json.Unmarshal(data, &epochData); err != nil {
		return nil, err
	}

	return epochData.Headers, nil
}

// LoadBlockRangeFromFile loads a specific range of blocks from epoch JSON file
// Example: LoadBlockRangeFromFile("epoch_1.json", 900, 950) loads blocks 900-950
func LoadBlockRangeFromFile(filename string, startBlock, endBlock uint64) ([]BlockHeaderJSON, error) {
	headers, err := LoadEpochFromFile(filename)
	if err != nil {
		return nil, err
	}

	// Filter headers in the specified range
	var filtered []BlockHeaderJSON
	for _, header := range headers {
		// Parse block number from hex string
		var blockNum uint64
		if _, err := fmt.Sscanf(header.Number, "0x%x", &blockNum); err != nil {
			continue
		}

		if blockNum >= startBlock && blockNum <= endBlock {
			filtered = append(filtered, header)
		}
	}

	return filtered, nil
}
