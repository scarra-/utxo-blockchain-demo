package network

import (
	"github.com/aadejanovs/blockchain-demo/blockchain"
)

func (s *Server) MineTx() {
	var txs []*blockchain.Transaction

	if s.Mempool.Len() <= 0 {
		s.Logger.Infow("mempool_empty")
		return
	}

	s.Logger.Infow("mining_block",
		"mempool_len", s.Mempool.Len(),
		"block_time", s.BlockTime,
	)

	for id := range s.Mempool.pool {
		tx, _ := s.Mempool.Get(id)
		if s.chain.VerifyTransaction(tx) {
			txs = append(txs, tx)
		}
	}

	if len(txs) == 0 {
		s.Logger.Errorw("all_transactions_invalid")
		return
	}

	cbTx := blockchain.CoinbaseTx(s.MinerAddress, "")
	txs = append(txs, cbTx)

	newBlock := s.chain.MineBlock(txs)
	UTXOSet := blockchain.UTXOSet{Blockchain: s.chain}
	UTXOSet.Reindex()

	s.Logger.Infow("new_block_mined",
		"hash", newBlock.GetHash(),
	)

	for _, tx := range txs {
		s.Mempool.Delete(tx.GetID())
	}

	s.PeersStorage.ForEach(func(peerAddr string) {
		s.client.SendBlockCreated(peerAddr, newBlock)
	})

	if s.Mempool.Len() > 0 {
		s.MineTx()
	}
}
