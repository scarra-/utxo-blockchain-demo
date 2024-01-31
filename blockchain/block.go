package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Block struct {
	Timestamp    int64
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
	Height       int
}

// DEBUG
func (b *Block) TxIds() []string {
	var ids []string

	b.SortTxs()

	for _, tx := range b.Transactions {
		ids = append(ids, tx.GetID())
	}

	return ids
}

// DEBUG
func (b *Block) TxInfo() map[string]string {
	info := make(map[string]string)

	for _, tx := range b.Transactions {
		txBytes, _ := json.Marshal(tx)
		info[tx.GetID()] = string(txBytes)
	}

	return info
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	b.SortTxs()

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	tree := NewMerkleTree(txHashes)

	return tree.MerkleRoot()
}

func (b *Block) JsonHashTransactions() []byte {
	var txHashes [][]byte

	b.SortTxs()

	for _, tx := range b.Transactions {
		serializedTx := tx.JsonSerialize()
		hash := sha256.Sum256(serializedTx)
		txHashes = append(txHashes, hash[:])
	}

	tree := NewMerkleTree(txHashes)

	return tree.MerkleRoot()
}

func (b *Block) SortTxs() {
	SortTxs(b.Transactions)
}

func CreateBlock(txs []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{
		Timestamp:    time.Now().Unix(),
		Hash:         []byte{},
		Transactions: txs,
		PrevHash:     prevHash,
		Nonce:        0,
		Height:       height,
	}

	block.SortTxs()

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (b *Block) GetHash() string {
	return fmt.Sprintf("%x", b.Hash)
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)
	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	Handle(err)

	block.SortTxs()

	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
