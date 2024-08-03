package dataflow

import "github.com/anon/execution-property-graph/dynamicEVM/types"

type context struct {
	callId          types.CallId
	addr            types.Address
	codeAddr        types.Address
	caller          types.Address
	mem             *memory
	sta             *stack
	callValue       word
	childCallValue  word
	childCallData   []*DataSource
	callData        []*DataSource
	retOffset       uint64
	retLength       uint64
	returnData      []*DataSource
	childReturnData []*DataSource

	storageWrites []*StorageLocation
	balanceWrites []*BalanceLocation
}
