package posv

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// TestCorePattern tests the exact pattern used in core/blockchain_test.go
func TestCorePattern(t *testing.T) {
	genesis := &core.Genesis{
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

	engine := ethash.NewFaker()

	t.Log("Step 1: Generate blocks")
	genDb, blocks, _ := core.GenerateChainWithGenesis(genesis, engine, 10, func(i int, b *core.BlockGen) {
		b.SetCoinbase(common.Address{0: byte(1), 19: byte(i)})
	})
	t.Logf("Generated %d blocks", len(blocks))

	t.Log("Step 2: Create blockchain with genDb")
	blockchain, err := core.NewBlockChain(genDb, core.DefaultCacheConfigWithScheme("hash"), genesis, nil, engine, vm.Config{}, nil)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}
	defer blockchain.Stop()
	t.Logf("Blockchain created, current: %d", blockchain.CurrentBlock().Number.Uint64())

	t.Log("Step 3: Insert blocks")
	n, err := blockchain.InsertChain(blocks)
	if err != nil {
		t.Fatalf("Insert failed at block %d: %v", n, err)
	}

	t.Logf("SUCCESS! Current block: %d", blockchain.CurrentBlock().Number.Uint64())
}
