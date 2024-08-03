package types

import (
	"math/big"

	"github.com/holiman/uint256"
)

const MaxUint64 = 1<<64 - 1

type Address interface {
	Hex() string
	Bytes() []byte
}

type Hash interface {
	Hex() string
	Bytes() []byte
}

type EVM interface {
}

type ScopeContext struct {
	Memory   Memory
	Stack    Stack
	Contract *Contract
}

type Memory interface {
	Data() []byte
}

type Stack interface {
	Data() []uint256.Int
	Manipulate(int, uint256.Int)
}

type Contract struct {
	Code []byte
	// The address of current executing contract
	// See geth core/vm/contract.go
	Address Address
}

type Header struct {
	Number     *big.Int
	Difficulty *big.Int
}

var (
	NullAddress        Address
	NativeTokenAddress Address

	BytesToAddress func([]byte) Address
	Bytes32ToHash  func([32]byte) Hash
	HexToBytes32   func(string) [32]byte
	HexToAddress   func(string) Address
	AddressToHex   func(Address) string

	IsPrecompile func(Address, *Header) bool
)
