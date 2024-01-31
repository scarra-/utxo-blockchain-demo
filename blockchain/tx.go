package blockchain

import (
	"bytes"
	"encoding/gob"

	"github.com/aadejanovs/blockchain-demo/wallet"
)

type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

type TxInputSort []TxInput

func (s TxInputSort) Len() int           { return len(s) }
func (s TxInputSort) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TxInputSort) Less(i, j int) bool { return string(s[i].ID) < string(s[j].ID) }

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputSort []TxOutput

func (s TxOutputSort) Len() int           { return len(s) }
func (s TxOutputSort) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TxOutputSort) Less(i, j int) bool { return string(s[i].PubKeyHash) < string(s[j].PubKeyHash) }

type TxOutputs struct {
	Outputs []TxOutput
}

func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{
		Value: value,
	}

	txo.Lock([]byte(address))

	return txo
}

func (outs TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(outs)
	Handle(err)

	return buffer.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	var outs TxOutputs
	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&outs)
	Handle(err)

	return outs
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Equal(lockingHash, pubKeyHash)
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}
