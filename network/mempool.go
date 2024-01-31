package network

import (
	"sync"

	"github.com/aadejanovs/blockchain-demo/blockchain"
	"go.uber.org/zap"
)

type Mempool struct {
	Logger *zap.SugaredLogger

	poolLock sync.RWMutex
	pool     map[string]*blockchain.Transaction
}

func NewMemPool(logger *zap.SugaredLogger) *Mempool {
	return &Mempool{
		Logger: logger,
		pool:   make(map[string]*blockchain.Transaction),
	}
}

func (m *Mempool) Get(key string) (*blockchain.Transaction, bool) {
	m.poolLock.RLock()
	defer m.poolLock.RUnlock()

	item, ok := m.pool[key]
	return item, ok
}

func (m *Mempool) Add(tx *blockchain.Transaction) {
	// m.poolLock.Lock()
	// defer m.poolLock.Unlock()

	_, ok := m.pool[tx.GetID()]

	if !ok {
		m.pool[tx.GetID()] = tx

		m.Logger.Infow("tx_added_to_mempool",
			"id", tx.GetID(),
			"pool_len", m.Len(),
			"pool_tx_ids", m.TxIds(),
		)
	} else {
		m.Logger.Infow("tx_already_exists_in_mempool",
			"id", tx.GetID(),
			"pool_len", m.Len(),
			"pool_tx_ids", m.TxIds(),
		)
	}
}

func (m *Mempool) Txs() []blockchain.Transaction {
	m.poolLock.RLock()
	defer m.poolLock.RUnlock()

	txs := []blockchain.Transaction{}

	for _, tx := range m.pool {
		txs = append(txs, *tx)
	}

	return txs
}

func (m *Mempool) TxIds() []string {
	m.poolLock.RLock()
	defer m.poolLock.RUnlock()

	txIds := []string{}

	for _, tx := range m.pool {
		txIds = append(txIds, tx.GetID())
	}

	return txIds
}

func (m *Mempool) Delete(key string) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()

	delete(m.pool, key)

	m.Logger.Infow("tx_removed_from_mempool",
		"id", key,
	)
}

func (m *Mempool) Len() int {
	return len(m.pool)
}

func (m *Mempool) ForEach(callback func(tx *blockchain.Transaction)) {
	m.poolLock.RLock()
	defer m.poolLock.RUnlock()

	for _, tx := range m.pool {
		callback(tx)
	}
}
