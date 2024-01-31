package cli

import (
	"fmt"
	"log"

	"github.com/aadejanovs/blockchain-demo/blockchain"
	"github.com/aadejanovs/blockchain-demo/wallet"
	"github.com/spf13/cobra"
)

var (
	createChainCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates new blockchain",
		Long:  `creates new blockchain and sends genesis reward to address`,
		Run:   createChain,
	}
)

func createChain(cmd *cobra.Command, args []string) {
	address, _ := cmd.Flags().GetString("addr")
	if !wallet.ValidateAddress(address) {
		log.Panic("Address not valid")
	}

	chain := blockchain.InitBlockchain(address, nodeID)

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")
	chain.Database.Close()
}
