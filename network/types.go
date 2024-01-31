package network

import "github.com/aadejanovs/blockchain-demo/blockchain"

type (
	Addr struct {
		AddrList []string
	}

	ChainBlockHashes struct {
		AddrFrom string
		Hashes   [][]byte
	}

	Block struct {
		AddrFrom string
		Block    []byte
	}

	BlockCreated struct {
		AddrFrom  string
		BlockHash []byte
	}

	GetBlock struct {
		AddrFrom string
		Hash     []byte
	}

	GetBlockByHeight struct {
		AddrFrom string
		Height   int
	}

	MemPoolTxs struct {
		AddrFrom string
		TxIds    []byte
	}

	GetBlocks struct {
		AddrFrom string
	}

	GetData struct {
		AddrFrom string
		Type     string
		ID       []byte
	}

	GetMempoolTxs struct {
		AddrFrom string
	}

	MempoolTxs struct {
		AddrFrom string
		Txs      []blockchain.Transaction
	}

	Inv struct {
		AddrFrom string
		Type     string
		Items    [][]byte
	}

	Tx struct {
		AddrFrom    string
		Transaction []byte
	}

	Version struct {
		Version    int
		BestHeight int
		AddrFrom   string
	}
)
