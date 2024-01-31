package blockchain

import (
	"bytes"
	"log"

	"github.com/cbergoon/merkletree"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{hash: hash}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(h.hash, other.(TxHash).hash)
	return equals, nil
}

func NewMerkleTree(data [][]byte) *merkletree.MerkleTree {
	var list []merkletree.Content

	for _, item := range data {
		list = append(list, TxHash{hash: item})
	}

	t, err := merkletree.NewTree(list)
	if err != nil {
		log.Fatal(err)
	}

	return t
}
