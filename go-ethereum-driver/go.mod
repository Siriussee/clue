module github.com/anon/execution-property-graph/go-ethereum-driver

go 1.18

replace github.com/anon/execution-property-graph/dynamicEVM => ../dynamicEVM

require (
	github.com/anon/execution-property-graph/dynamicEVM v0.0.0
	github.com/ethereum/go-ethereum v1.10.23
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/holiman/uint256 v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
)
