package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "Genesis data"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
	Logger   *zap.SugaredLogger
}

func DBExists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

func SetupLogger(nodeId string) (*zap.SugaredLogger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = nil
	config.EncoderConfig.TimeKey = ""
	config.InitialFields = map[string]interface{}{
		"app_name": fmt.Sprintf("node-%s", nodeId),
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

func ContinueBlockchain(nodeId string) *Blockchain {
	path := fmt.Sprintf(dbPath, nodeId)
	if !DBExists(path) {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(path)
	opts.EventLogging = false
	opts.Logger = nil

	db, err := OpenDB(path, opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		return err
	})
	Handle(err)

	logger, err := SetupLogger(nodeId)
	Handle(err)

	return &Blockchain{
		LastHash: lastHash,
		Database: db,
		Logger:   logger,
	}
}

func InitBlockchain(address, nodeId string) *Blockchain {
	var lastHash []byte
	path := fmt.Sprintf(dbPath, nodeId)
	if DBExists(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(path)
	opts.EventLogging = false
	opts.Logger = nil

	db, err := OpenDB(path, opts)
	Handle(err)

	logger, err := SetupLogger(nodeId)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})

	Handle(err)

	return &Blockchain{
		LastHash: lastHash,
		Database: db,
		Logger:   logger,
	}
}

func (chain *Blockchain) ValidateBlock(block *Block) error {
	chain.Logger.Infow("block_validation_started", "hash", block.GetHash())

	lastBlock, err := chain.GetLastBlock()
	if err != nil {
		return err
	}

	if block.Height != lastBlock.Height+1 {
		chain.Logger.Warnw("new_block_height_invalid",
			"new_block_hash", fmt.Sprintf("%x", block.Hash),
			"new_block_height", block.Height,
			"last_block_hash", lastBlock.Height,
		)
	}

	if !bytes.Equal(block.PrevHash, lastBlock.Hash) {
		chain.Logger.Warnw("block_prevhash_doesnt_match",
			"last_block_hash", fmt.Sprintf("%x", lastBlock.Hash),
			"new_block_prev_hash", fmt.Sprintf("%x", block.PrevHash),
		)
		return fmt.Errorf("block prev hash %x doesnt match last block hash %x", block.PrevHash, lastBlock.Hash)
	}

	for _, tx := range block.Transactions {
		if !chain.VerifyTransaction(tx) {
			chain.Logger.Warnw("invalid_transaction",
				"block_hash", block.GetHash(),
				"tx", tx,
			)

			return fmt.Errorf("invalid block %s tx %s", block.GetHash(), tx.GetID())
		}
	}

	pow := NewProof(block)
	if !pow.Validate() {
		chain.Logger.Warnw("block_pow_validation_failed", "hash", block.GetHash())
		return fmt.Errorf("block %s pow validation failed", block.GetHash())
	}

	chain.Logger.Infow("block_validation_completed", "hash", block.GetHash())

	return nil
}

func (chain *Blockchain) GetLastBlock() (*Block, error) {
	var block *Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.ValueCopy(nil)

		if item, err := txn.Get(lastHash); err != nil {
			return errors.New("Block not found")
		} else {
			blockData, _ := item.ValueCopy(nil)
			block = Deserialize(blockData)
		}

		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *Blockchain) AddBlock(block *Block) error {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(block.Hash)

		if err == nil {
			chain.Logger.Infow("block_already_exists",
				"hash", fmt.Sprintf("%x", block.Hash),
				"height", block.Height,
			)
			return nil
		}

		if err := chain.ValidateBlock(block); err != nil {
			return err
		}

		blockData := block.Serialize()
		err = txn.Set(block.Hash, blockData)
		if err != nil {
			chain.Logger.Panicw("adding_block_to_db_failed",
				"hash", fmt.Sprintf("%x", block.Hash),
				"height", block.Height,
				"error", err,
			)
		}

		err = txn.Set([]byte("lh"), block.Hash)
		Handle(err)
		chain.LastHash = block.Hash

		chain.Logger.Infow("added_new_block",
			"hash", fmt.Sprintf("%x", block.Hash),
			"height", block.Height,
		)

		return nil
	})

	return err
}

func (chain *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block not found")
		} else {
			blockData, _ := item.ValueCopy(nil)
			block = *Deserialize(blockData)
		}

		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *Blockchain) GetBlockByHeight(height int) (*Block, error) {
	var block *Block

	iter := chain.Iterator()

	for {
		block = iter.Next()

		if block.Height == height {
			return block, nil
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return nil, fmt.Errorf("block at height - %d not found", height)
}

func (chain *Blockchain) BlockExists(blockHash []byte) bool {
	err := chain.Database.View(func(txn *badger.Txn) error {
		_, err := txn.Get(blockHash)
		return err
	})

	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			chain.Logger.Errorf("error_while_getting_block_from_db",
				"error", err,
			)
		}

		return false
	}

	return true
}

func (chain *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.ValueCopy(nil)

		item, _ = txn.Get(lastHash)
		blockData, _ := item.ValueCopy(nil)
		lastBlock = *Deserialize(blockData)
		return nil
	})
	Handle(err)

	return lastBlock.Height
}

func (chain *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int
	for _, tx := range transactions {
		chain.Logger.Infow("transaction_verification_started",
			"tx_id", fmt.Sprintf("%x", tx.ID),
		)
		if !chain.VerifyTransaction(tx) {
			chain.Logger.Panicw("invalid_transaction",
				"transaction", tx,
			)
		}
		chain.Logger.Infow("transaction_verified",
			"transaction", tx,
		)
	}

	chain.Logger.Infof("all_transactions_verified")

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		item.Value(
			func(val []byte) error {
				lastHash = val
				return nil
			},
		)

		item, err = txn.Get(lastHash)
		Handle(err)

		var lastBlockData []byte
		err = item.Value(
			func(val []byte) error {
				lastBlockData = val
				return nil
			},
		)

		lastBlock := Deserialize(lastBlockData)
		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)

	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	chain.Logger.Infow("new_block_mined",
		"block_hash", newBlock.GetHash(),
		"block_height", newBlock.Height,
	)

	return newBlock
}

func (chain *Blockchain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := tx.GetID()

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ed25519.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		bc.Logger.Infow("skipping_coinbase_transaction_verification",
			"tx_id", tx.GetID(),
		)
		return true
	}

	prevTXs := make(map[string]Transaction)

	bc.Logger.Infow("verifying_tx_inputs",
		"tx_id", tx.GetID(),
	)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)

		bc.Logger.Infow("got_db_result",
			"tx_id", tx.GetID(),
			"err", err,
		)

		if err != nil {
			bc.Logger.Panicw("transaction_verification_failed",
				"error", err,
			)
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX

		bc.Logger.Infow("added_tx_to_prev_txs",
			"tx_id", tx.GetID(),
		)
	}
	res := tx.Verify(prevTXs)

	bc.Logger.Infow("tx_verification_result",
		"tx_id", tx.GetID(),
		"result", res,
	)

	return res

}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func OpenDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
