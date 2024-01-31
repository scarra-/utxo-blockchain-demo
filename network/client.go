package network

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/aadejanovs/blockchain-demo/blockchain"
	"go.uber.org/zap"
)

type ClientSettings struct {
	Version       int
	Protocol      string
	CommandLength int
}

type Client struct {
	ClientSettings
	Logger      *zap.SugaredLogger
	nodeAddress string
}

func NewClient(logger *zap.SugaredLogger, nodeAddress string) *Client {
	return &Client{
		nodeAddress: nodeAddress,
		Logger:      logger,
		ClientSettings: ClientSettings{
			Version:       1,
			Protocol:      "tcp",
			CommandLength: 32,
		},
	}
}

func (c *Client) SendVersion(addr string, chain *blockchain.Blockchain) {
	bestHeight := chain.GetBestHeight()
	payload := GobEncode(Version{Version: c.Version, BestHeight: bestHeight, AddrFrom: c.nodeAddress})
	request := append(c.MsgNameToBytes(msgVersion), payload...)

	c.Logger.Infow("sending_version_to_peer",
		"peer_addr", addr,
		"host_addr", c.nodeAddress,
		"host_version", c.Version,
		"host_height", bestHeight,
	)

	c.SendData(addr, request)
}

func (c *Client) SendAddresses(addr string, peerAddresses []string) {
	nodes := Addr{AddrList: peerAddresses}
	nodes.AddrList = append(nodes.AddrList, c.nodeAddress)
	payload := GobEncode(nodes)
	request := append(c.MsgNameToBytes(msgAddresses), payload...)

	c.Logger.Infow("sending_addresses_to_peer",
		"peer_addr", addr,
	)

	c.SendData(addr, request)
}

func (c *Client) SendBlockCreated(addr string, b *blockchain.Block) {
	data := BlockCreated{AddrFrom: c.nodeAddress, BlockHash: b.Hash}
	payload := GobEncode(data)
	request := append(c.MsgNameToBytes(msgBlockCreated), payload...)

	c.Logger.Infow("sending_block_created_to_peer",
		"peer_addr", addr,
		"block_hash", b.GetHash(),
	)

	c.SendData(addr, request)
}

func (c *Client) SendBlock(addr string, b *blockchain.Block) {
	data := Block{AddrFrom: c.nodeAddress, Block: b.Serialize()}
	payload := GobEncode(data)
	request := append(c.MsgNameToBytes(msgBlock), payload...)

	c.Logger.Infow("sending_block_to_peer",
		"peer_addr", addr,
		"block_hash", b.GetHash(),
	)

	c.SendData(addr, request)
}

func (c *Client) GetNextBlock(addr string, currentHeight int) {
	height := currentHeight + 1
	data := GetBlockByHeight{AddrFrom: c.nodeAddress, Height: height}
	payload := GobEncode(data)
	request := append(c.MsgNameToBytes(msgGetBlockByHeight), payload...)

	c.Logger.Infow("sending_get_block_by_height_from_peer",
		"peer_addr", addr,
		"block_height", height,
	)

	c.SendData(addr, request)
}

func (c *Client) SendGetBlock(addr string, hash []byte) {
	payload := GobEncode(GetBlock{AddrFrom: c.nodeAddress, Hash: hash})
	request := append(c.MsgNameToBytes(msgGetBlock), payload...)

	c.Logger.Infow("requesting_block_from_peer",
		"peer_addr", addr,
		"block_hash", fmt.Sprintf("%x", hash),
	)

	c.SendData(addr, request)
}

// Called when we create TX from CLI. Sends TX to founding node that adds it to mempool.
// If it was added to mempool - node propogates tx to its peers.
func (c *Client) SendTx(addr string, tx *blockchain.Transaction) {
	data := Tx{AddrFrom: c.nodeAddress, Transaction: tx.Serialize()}
	payload := GobEncode(data)
	request := append(c.MsgNameToBytes(msgTx), payload...)

	c.Logger.Infow("sending_tx_to_peer",
		"peer_addr", addr,
		"tx_id", tx.GetID(),
	)

	c.SendData(addr, request)
}

func (c *Client) SendGetMempoolTxs(addr string) {
	payload := GobEncode(GetMempoolTxs{AddrFrom: c.nodeAddress})
	request := append(c.MsgNameToBytes(msgGetMempoolTxs), payload...)

	c.Logger.Infow("requesting_mempool_txs_from_peer",
		"peer_addr", addr,
	)

	c.SendData(addr, request)
}

func (c *Client) SendMempoolTxs(addr string, txs []blockchain.Transaction) {
	payload := GobEncode(MempoolTxs{AddrFrom: c.nodeAddress, Txs: txs})
	request := append(c.MsgNameToBytes(msgMempoolTxs), payload...)

	c.Logger.Infow("sending_mempool_txs_to_peer",
		"peer_addr", addr,
	)

	c.SendData(addr, request)
}

func (c *Client) SendData(addr string, data []byte) error {
	conn, err := net.Dial(c.Protocol, addr)

	if err != nil {
		c.Logger.Errorf("connection_to_peer_failed",
			"peer_addr", addr,
			"error", err,
		)
		return nil
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		c.Logger.Errorf("sending_data_to_peer_failed",
			"peer_addr", addr,
			"error", err,
		)
		log.Panic(err)
	}

	return nil
}

func (c *Client) MsgNameToBytes(cmd string) []byte {
	bytes := make([]byte, c.CommandLength)

	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return bytes[:]
}
