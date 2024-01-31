package cli

import (
	"fmt"
	"log"

	"github.com/aadejanovs/blockchain-demo/network"
	"github.com/aadejanovs/blockchain-demo/wallet"
	"github.com/spf13/cobra"
)

var (
	startNodeCmd = &cobra.Command{
		Use:   "start",
		Short: "Start node",
		Long:  `Start a node with ID specified in NODE_ID env. var. -miner enables mining`,
		Run:   startNode,
	}
)

func startNode(cmd *cobra.Command, args []string) {
	fmt.Printf("Starting node %s\n", nodeID)

	minerAddress, _ := cmd.Flags().GetString("miner")

	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address")
		}
	}

	server := network.NewServer(nodeID, minerAddress)
	server.Start()
}
