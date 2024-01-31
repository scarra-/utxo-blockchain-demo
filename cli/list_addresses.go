package cli

import (
	"fmt"

	"github.com/aadejanovs/blockchain-demo/wallet"
	"github.com/spf13/cobra"
)

var (
	listAddressesCmd = &cobra.Command{
		Use:   "addr",
		Short: "Lists the addresses in our wallet file",
		Long:  "Lists the addresses in our wallet file",
		Run:   listAddresses,
	}
)

func listAddresses(cmd *cobra.Command, args []string) {
	wallets, err := wallet.Load(nodeID)
	fmt.Println(err)
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
