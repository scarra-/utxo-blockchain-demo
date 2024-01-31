package cli

import (
	"fmt"
	"log"

	"github.com/aadejanovs/blockchain-demo/blockchain"
	"github.com/aadejanovs/blockchain-demo/wallet"
	"github.com/spf13/cobra"
)

var (
	getBalanceCmd = &cobra.Command{
		Use:   "balance",
		Short: "Get the balance for an address",
		Long:  `Get the balance for an address`,
		Run:   getBalance,
	}
)

func getBalance(cmd *cobra.Command, args []string) {
	address, _ := cmd.Flags().GetString("addr")

	if !wallet.ValidateAddress(address) {
		log.Panic("Address not valid")
	}

	chain := blockchain.ContinueBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}
