## UTXO based blockhain in Go

<img width="1000" alt="Screenshot 2024-01-31 at 18 54 02" src="https://github.com/scarra-/utxo-blockchain-demo/assets/10938706/c133c00f-da84-4a27-80d2-13ff11742032">

### Available commands:

- `./bin/chain create` Initialize new chain. Node identifier is picked from `NODE_ID` env variable.
- `./bin/chain start --miner={true/false}` Start node.
- `./bin/chain reindex` Reindex UTXO database
- `./bin/chain create-wallet` Create wallet
- `./bin/chain addr` List local wallet addresses
- `./bin/chain balance --addr {wallet_address}` See address balance
- `./bin/chain send -from {from_addr} -to {to_addr} -amount {amount}` Send transaction
- `./bin/chain print` Print local chain with all blocks and transactions


### Libraries used

- `spf13/cobra` CLI application
- `uber-go/zap` Logger
- `dgraph-io/badger` BadgerDB is an embeddable, persistent and fast key-value (KV) database
- `cbergoon/merkletree` Library to generate Merkle tree
- `vrecan/death/v3` Library to catch signals that end your application
- `mr-tron/base58` Base58 encoder. Excludes characters that look similar like: O, 0, l, I.
