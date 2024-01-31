## UTXO based blockhain in Go

### Available commands:

Initialize new chain. Node identifier is picked from `NODE_ID` env variable:
- `./bin/chain create`

Start node:
- `./bin/chain start --miner={true/false}`

Reindex UTXO database:
- `./bin/chain reindex`

Create wallet:
- `./bin/chain create-wallet`

List local wallet addresses:
- `./bin/chain addr`

See address balance:
- `./bin/chain balance --addr {wallet_address}`

Send transaction:
- `./bin/chain send -from {from_addr} -to {to_addr} -amount {amount}`

Print local chain with all blocks and transactions:
- `./bin/chain print`


### Libraries used

- `spf13/cobra` CLI application
- `uber-go/zap` Logger
- `dgraph-io/badger` BadgerDB is an embeddable, persistent and fast key-value (KV) database
- `cbergoon/merkletree` Library to generate Merkle tree
- `vrecan/death/v3` Library to catch signals that end your application
- `mr-tron/base58` Base58 encoder. Excludes characters that look similar like: O, 0, l, I.
