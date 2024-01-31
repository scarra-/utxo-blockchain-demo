package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aadejanovs/blockchain-demo/wallet"
)

type Transaction struct {
	ID        []byte
	Inputs    []TxInput
	Outputs   []TxOutput
	Timestamp int64
}

func (tx *Transaction) GetID() string {
	return hex.EncodeToString(tx.ID)
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(tx); err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx Transaction) JsonSerialize() []byte {
	result, _ := json.Marshal(tx)

	return result
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	Handle(err)

	return transaction
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TxInput{
		ID:        nil,
		Out:       -1,
		Signature: nil,
		// Non existing pub key
		PubKey: []byte(data),
	}

	txout := NewTXOutput(20, to)

	tx := Transaction{
		ID:        nil,
		Inputs:    []TxInput{txin},
		Outputs:   []TxOutput{*txout},
		Timestamp: time.Now().UnixNano(),
	}

	tx.ID = tx.Hash()

	return &tx
}

func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(w.PublicKeyBytes())
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{
				ID:        txID,
				Out:       out,
				Signature: nil,
				PubKey:    w.PublicKeyBytes(),
			}

			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, string(w.Address())))
	}

	tx := Transaction{
		ID:        nil,
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now().UnixNano(),
	}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx *Transaction) Sign(privKey ed25519.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash

		signature := ed25519.Sign(privKey, []byte(fmt.Sprintf("%x", txCopy)))

		tx.Inputs[inId].Signature = signature
		txCopy.Inputs[inId].PubKey = nil
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{
			ID:        in.ID,
			Out:       in.Out,
			Signature: nil,
			PubKey:    nil,
		})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      out.Value,
			PubKeyHash: out.PubKeyHash,
		})
	}

	txCopy := Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}

	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range tx.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]

		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash

		result := ed25519.Verify(in.PubKey, []byte(fmt.Sprintf("%x", txCopy)), in.Signature)

		if !result {
			return false
		}

		txCopy.Inputs[inId].PubKey = nil
	}

	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

type TxList struct {
	txs []*Transaction
}

func SortTxs(txs []*Transaction) {
	txList := &TxList{txs}
	sort.Sort(txList)
}

func (s *TxList) Len() int           { return len(s.txs) }
func (s *TxList) Swap(i, j int)      { s.txs[i], s.txs[j] = s.txs[j], s.txs[i] }
func (s *TxList) Less(i, j int) bool { return s.txs[i].Timestamp < s.txs[j].Timestamp }
