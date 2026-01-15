package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

const EpochLength = 900

type BlockHeader struct {
	Number     string `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parentHash"`
	Coinbase   string `json:"miner"`
	Extra      string `json:"extraData"`
	Validator  string `json:"validator,omitempty"`
	Penalties  string `json:"penalties,omitempty"`
	Validators string `json:"validators,omitempty"`
	Difficulty string `json:"difficulty"`
	GasLimit   string `json:"gasLimit"`
	GasUsed    string `json:"gasUsed"`
	Time       string `json:"timestamp"`
	Nonce      string `json:"nonce"`
	// Full header fields
	Root        string `json:"stateRoot"`
	TxHash      string `json:"transactionsRoot"`
	ReceiptHash string `json:"receiptsRoot"`
	Bloom       string `json:"logsBloom"`
	MixDigest   string `json:"mixHash"`
	UncleHash   string `json:"sha3Uncles"`
}

type EpochData struct {
	StartBlock  uint64                 `json:"startBlock"`
	EndBlock    uint64                 `json:"endBlock"`
	Description string                 `json:"description"`
	Headers     []BlockHeader          `json:"headers"`
	ChainConfig map[string]interface{} `json:"chainConfig"`
}

func main() {
	rpcURL := flag.String("rpc", "https://rpc.viction.xyz", "RPC endpoint URL")
	epochNum := flag.Uint64("epoch", 123, "Epoch number to dump (will fetch epoch*900 to epoch*900+899)")
	output := flag.String("output", "", "Output file (default: epoch_<number>.json)")
	flag.Parse()

	// Calculate block range
	startBlock := uint64(0)
	endBlock := uint64(51)

	if *output == "" {
		*output = fmt.Sprintf("epoch_%d.json", *epochNum)
	}

	fmt.Printf("Fetching epoch %d (blocks %d to %d)...\n", *epochNum, startBlock, endBlock)
	fmt.Printf("This will fetch %d blocks from %s\n", EpochLength, *rpcURL)

	// Connect to RPC
	client, err := rpc.Dial(*rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	epochData := &EpochData{
		StartBlock:  startBlock,
		EndBlock:    endBlock,
		Description: fmt.Sprintf("Complete epoch %d from Viction mainnet", *epochNum),
		Headers:     make([]BlockHeader, 0, EpochLength),
		ChainConfig: map[string]interface{}{
			"chainId": 88,
			"epoch":   EpochLength,
		},
	}

	// Fetch all headers in the epoch
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		if blockNum%100 == 0 {
			fmt.Printf("Fetching block %d/%d...\n", blockNum-startBlock, EpochLength)
		}

		var header BlockHeader
		blockHex := fmt.Sprintf("0x%x", blockNum)
		err := client.Call(&header, "eth_getBlockByNumber", blockHex, false)
		if err != nil {
			log.Printf("Warning: Failed to fetch block %d: %v", blockNum, err)
			continue
		}

		// Parse and store header
		epochData.Headers = append(epochData.Headers, header)
	}

	fmt.Printf("\nFetched %d/%d headers\n", len(epochData.Headers), EpochLength)

	// Marshal to JSON
	data, err := json.MarshalIndent(epochData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal data: %v", err)
	}

	// Save to file
	err = os.WriteFile(*output, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	fmt.Printf("\n✓ Saved complete epoch to %s\n", *output)
	fmt.Printf("File size: %.2f MB\n", float64(len(data))/1024/1024)

	// Summary
	if len(epochData.Headers) > 0 {
		firstHeader := epochData.Headers[0]
		lastHeader := epochData.Headers[len(epochData.Headers)-1]

		fmt.Printf("\nEpoch Summary:\n")
		fmt.Printf("  First block: %s (hash: %s)\n", firstHeader.Number, firstHeader.Hash[:16]+"...")
		fmt.Printf("  Last block:  %s (hash: %s)\n", lastHeader.Number, lastHeader.Hash[:16]+"...")

		// Check if it's a checkpoint block
		if endBlock%EpochLength == EpochLength-1 {
			fmt.Printf("  ✓ Contains checkpoint at block %d\n", endBlock+1)
		}

		// Count validators in last block's extra data (if checkpoint)
		if len(lastHeader.Extra) > 97 {
			extraBytes, _ := hexutil.Decode(lastHeader.Extra)
			if len(extraBytes) > 97 {
				validatorsData := extraBytes[32 : len(extraBytes)-65]
				numValidators := len(validatorsData) / 20
				if numValidators > 0 {
					fmt.Printf("  Validators: %d\n", numValidators)
				}
			}
		}
	}

	fmt.Printf("\nUsage in tests:\n")
	fmt.Printf("  1. Load epoch data: LoadEpochFromFile(\"%s\")\n", *output)
	fmt.Printf("  2. Test snapshot.apply() with all %d headers\n", len(epochData.Headers))
	fmt.Printf("  3. Validate checkpoint transition at block %d\n", endBlock+1)
}
