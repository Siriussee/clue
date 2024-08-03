package calltrace

import (
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

type CallTrace struct {
	ID   types.CallId
	From types.Address
	To   types.Address
	Type vm.OpCode

	Parent *CallTrace

	Flows
}

type CallTraces []*CallTrace
