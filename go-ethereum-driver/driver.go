package driver

import (
	"encoding/hex"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func init() {
	types.NullAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
	types.NativeTokenAddress = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")

	types.BytesToAddress = func(b []byte) types.Address {
		return common.BytesToAddress(b)
	}
	types.Bytes32ToHash = func(b [32]byte) types.Hash {
		return common.BytesToHash(b[:])
	}
	types.HexToAddress = func(s string) types.Address {
		return common.HexToAddress(s)
	}
	types.IsPrecompile = func(addr types.Address, header *types.Header) bool {
		var precompiles map[common.Address]vm.PrecompiledContract
		chainRules := params.MainnetChainConfig.Rules(header.Number, header.Difficulty.Cmp(common.Big0) == 0)
		switch {
		case chainRules.IsBerlin:
			precompiles = vm.PrecompiledContractsBerlin
		case chainRules.IsIstanbul:
			precompiles = vm.PrecompiledContractsIstanbul
		case chainRules.IsByzantium:
			precompiles = vm.PrecompiledContractsByzantium
		default:
			precompiles = vm.PrecompiledContractsHomestead
		}
		_, ok := precompiles[common.BytesToAddress(addr.Bytes())]
		return ok
	}
	types.HexToBytes32 = func(s string) [32]byte {
		return common.HexToHash(s)
	}
	types.AddressToHex = func(a types.Address) string {
		return "0x" + hex.EncodeToString(a.Bytes())
	}
}
