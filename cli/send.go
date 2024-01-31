package cli

import (
	"fmt"
	"log"

	"github.com/aadejanovs/blockchain-demo/blockchain"
	"github.com/aadejanovs/blockchain-demo/network"
	"github.com/aadejanovs/blockchain-demo/wallet"
	"github.com/spf13/cobra"
)

var (
	sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send coins to address.",
		Long:  `send -from FROM -to TO -amount AMOUNT -mine - Send amount of coins.`,
		Run:   send,
	}
)

func send(cmd *cobra.Command, args []string) {
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	amount, _ := cmd.Flags().GetInt("amount")

	if !wallet.ValidateAddress(to) {
		log.Panic("Address not valid")
	}
	if !wallet.ValidateAddress(from) {
		log.Panic("Address not valid")
	}

	chain := blockchain.ContinueBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	chain.Logger.Infow("loading_wallets")
	wallets, err := wallet.Load(nodeID)
	if err != nil {
		chain.Logger.Panicw("error_loading_wallets",
			"error", err,
		)
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)

	chain.Logger.Infow("created_new_transaction",
		"tx_id", tx.GetID(),
		"from_addr", from,
		"to_addr", to,
	)

	client := network.NewClient(chain.Logger, fmt.Sprintf("localhost:%s", nodeID))
	client.SendTx("localhost:3000", tx)

	chain.Logger.Infow("new_tx_sent_to_founding_node",
		"tx_id", tx.GetID(),
		"from_addr", from,
		"to_addr", to,
	)
}
