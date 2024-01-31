package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	nodeID string

	rootCmd = &cobra.Command{
		Use:   "chain-cli",
		Short: "UTXO based blockchain demo application",
		Long:  `UTXO based blockchain demo application`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	nodeID = os.Getenv("NODE_ID")

	getBalanceCmd.Flags().StringP("addr", "a", "", "Specify the address for balance")
	getBalanceCmd.MarkFlagRequired("addr")
	rootCmd.AddCommand(getBalanceCmd)

	createChainCmd.Flags().StringP("addr", "a", "", "Specify the address for block reward")
	createChainCmd.MarkFlagRequired("addr")
	rootCmd.AddCommand(createChainCmd)

	sendCmd.Flags().StringP("from", "f", "", "Specify the from address")
	sendCmd.MarkFlagRequired("to")
	sendCmd.Flags().StringP("to", "t", "", "Specify the target address")
	sendCmd.MarkFlagRequired("to")
	sendCmd.Flags().IntP("amount", "a", 5, "Specify amount")
	sendCmd.MarkFlagRequired("amount")
	sendCmd.Flags().BoolP("mine", "m", false, "Mine now")
	rootCmd.AddCommand(sendCmd)

	startNodeCmd.Flags().StringP("miner", "m", "", "Specify the address for mining rewards")
	rootCmd.AddCommand(startNodeCmd)

	rootCmd.AddCommand(printChainCmd)
	rootCmd.AddCommand(listAddressesCmd)
	rootCmd.AddCommand(reindexUTXOCmd)
	rootCmd.AddCommand(createWalletCmd)
}
