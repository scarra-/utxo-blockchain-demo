package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

const Difficulty = 18

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)

	target.Lsh(target, uint(256-Difficulty))

	pow := &ProofOfWork{
		Block:  b,
		Target: target,
	}

	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.JsonHashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)

	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			// pow.Logger.Infow("block_mined_with_params",
			// 	"prev_hash", pow.Block.PrevHash,
			// 	"txs_hash", pow.Block.HashTransactions(),
			// 	"nonce", ToHex(int64(nonce)),
			// 	"difficulty", ToHex(int64(Difficulty)),
			// 	"txs_len", len(pow.Block.Transactions),
			// 	"txs", pow.Block.TxIds(),
			// 	"tx_info", pow.Block.TxInfo(),
			// )

			break
		} else {
			nonce++
		}

	}

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	// pow.Logger.Infow("validating_block_pow_with_params",
	// 	"prev_hash", pow.Block.PrevHash,
	// 	"txs_hash", pow.Block.HashTransactions(),
	// 	"nonce", ToHex(int64(pow.Block.Nonce)),
	// 	"difficulty", ToHex(int64(Difficulty)),
	// 	"txs_len", len(pow.Block.Transactions),
	// 	"txs", pow.Block.TxIds(),
	// 	"tx_info", pow.Block.TxInfo(),
	// )

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
