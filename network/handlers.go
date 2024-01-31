package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/aadejanovs/blockchain-demo/blockchain"
)

func (s *Server) HandleVersion(request []byte) {
	payload := DecodeRequest[Version](request, s.MsgNameLength)

	bestHeight := s.chain.GetBestHeight()
	otherHeight := payload.BestHeight

	s.Logger.Infow("received_version_request",
		"requester_height", payload.BestHeight,
		"requester_addr", payload.AddrFrom,
		"requester_version", payload.Version,
		"host_height", bestHeight,
		"host_addr", s.NodeAddress,
		"host_version", s.Version,
	)

	if !s.PeersStorage.PeerExists(payload.AddrFrom) {
		s.PeersStorage.Add(payload.AddrFrom)
	}

	if bestHeight < otherHeight {
		s.client.GetNextBlock(payload.AddrFrom, bestHeight)
	} else if bestHeight > otherHeight {
		s.client.SendVersion(payload.AddrFrom, s.chain)
	}
}

func (s *Server) HandleAddresses(request []byte) {
	payload := DecodeRequest[Addr](request, s.MsgNameLength)

	for _, addr := range payload.AddrList {
		s.PeersStorage.Add(addr)
	}

	s.Logger.Infow("received_addr_message",
		"peers_len", len(payload.AddrList),
		"known_peers", s.PeersStorage.Len(),
	)

	s.PeersStorage.ForEach(func(peerAddr string) {
		s.client.SendVersion(peerAddr, s.chain)
	})
}

func (s *Server) HandleBlockCreated(request []byte) {
	payload := DecodeRequest[BlockCreated](request, s.MsgNameLength)

	s.Logger.Infow("received_block_created_message",
		"addr_from", payload.AddrFrom,
		"block_hash", payload.BlockHash,
	)

	if !s.chain.BlockExists(payload.BlockHash) {
		s.client.SendGetBlock(payload.AddrFrom, payload.BlockHash)
	}
}

func (s *Server) HandleGetBlock(request []byte) {
	payload := DecodeRequest[GetBlock](request, s.MsgNameLength)

	s.Logger.Infow("received_get_block_query",
		"addr_from", payload.AddrFrom,
		"block_hash", fmt.Sprintf("%x", payload.Hash),
	)

	block, err := s.chain.GetBlock(payload.Hash)
	if err != nil {
		return
	}

	s.client.SendBlock(payload.AddrFrom, &block)
}

func (s *Server) HandleBlockByHeightRequested(request []byte) {
	payload := DecodeRequest[GetBlockByHeight](request, s.MsgNameLength)

	s.Logger.Infow("received_get_block_by_height_query",
		"addr_from", payload.AddrFrom,
		"block_height", fmt.Sprintf("%x", payload.Height),
	)

	block, err := s.chain.GetBlockByHeight(payload.Height)
	if err != nil {
		return
	}

	s.client.SendBlock(payload.AddrFrom, block)
}

func (s *Server) HandleBlock(request []byte) {
	payload := DecodeRequest[Block](request, s.MsgNameLength)

	blockData := payload.Block
	block := blockchain.Deserialize(blockData)

	s.Logger.Infow("received_block_message",
		"addr_from", payload.AddrFrom,
		"block_hash", block.GetHash(),
		"block_height", block.Height,
		"block_tx_len", len(block.Transactions),
	)

	if !s.chain.BlockExists(block.Hash) {
		err := s.chain.AddBlock(block)
		if err == nil {
			for _, txID := range block.TxIds() {
				UTXOSet := blockchain.UTXOSet{Blockchain: s.chain}
				UTXOSet.Reindex()

				s.Mempool.Delete(txID)
			}
		}

		s.client.SendVersion(s.NodeAddress, s.chain)
	}
}

func (s *Server) HandleTx(request []byte) {
	payload := DecodeRequest[Tx](request, s.MsgNameLength)

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)

	s.Logger.Infow("received_tx_message",
		"addr_from", payload.AddrFrom,
		"tx_id", tx.GetID(),
		"mempool_len", s.Mempool.Len(),
	)

	_, txInMempool := s.Mempool.Get(tx.GetID())

	if !txInMempool {
		s.Mempool.Add(&tx)

		s.PeersStorage.ForEach(func(peerAddr string) {
			if peerAddr != payload.AddrFrom {
				s.client.SendTx(peerAddr, &tx)
			}
		})
	}
}

func (s *Server) HandleGetMempoolTxs(request []byte) {
	payload := DecodeRequest[GetMempoolTxs](request, s.MsgNameLength)

	s.Logger.Infow("received_get_mempool_txs_message",
		"addr_from", payload.AddrFrom,
	)

	s.Mempool.ForEach(func(tx *blockchain.Transaction) {
		go s.client.SendTx(payload.AddrFrom, tx)
	})
}

func DecodeRequest[T any](request []byte, cmdLen int) T {
	var buff bytes.Buffer
	var payload T

	buff.Write(request[cmdLen:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	return payload
}
