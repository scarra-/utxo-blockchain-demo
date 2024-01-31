package cli

import (
	"fmt"
	"strconv"

	"github.com/aadejanovs/blockchain-demo/blockchain"
	"github.com/spf13/cobra"
)

var (
	printChainCmd = &cobra.Command{
		Use:   "print",
		Short: "Prints the blocks in the chain",
		Long:  `Prints the blocks in the chain`,
		Run:   printChain,
	}
)

func printChain(cmd *cobra.Command, args []string) {
	chain := blockchain.ContinueBlockchain(nodeID)
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}
