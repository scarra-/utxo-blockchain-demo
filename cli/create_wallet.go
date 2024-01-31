package cli

import (
	"fmt"

	"github.com/aadejanovs/blockchain-demo/wallet"
	"github.com/spf13/cobra"
)

var (
	createWalletCmd = &cobra.Command{
		Use:   "create-wallet",
		Short: "Creates new wallet",
		Long:  "Creates new node wallet",
		Run:   createWallet,
	}
)

func createWallet(cmd *cobra.Command, args []string) {
	wallets, _ := wallet.Load(nodeID)
	address := wallets.AddWallet()
	wallets.SaveFile(nodeID)

	fmt.Printf("New address: %s\n", address)
}
