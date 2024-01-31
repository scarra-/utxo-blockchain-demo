package cli

import (
	"github.com/aadejanovs/blockchain-demo/blockchain"
	"github.com/spf13/cobra"
)

var (
	reindexUTXOCmd = &cobra.Command{
		Use:   "reindex",
		Short: "Rebuilds the UTXO set",
		Long:  `Rebuilds the UTXO set`,
		Run:   reindexUTXO,
	}
)

func reindexUTXO(cmd *cobra.Command, args []string) {
	chain := blockchain.ContinueBlockchain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	chain.Logger.Infow("utxo_set_reindexed",
		"tx_count", UTXOSet.CountTransactions(),
	)
}
