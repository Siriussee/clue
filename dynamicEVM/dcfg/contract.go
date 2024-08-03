package dcfg

import (
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

type ContractNode struct {
	callId        types.CallId
	addr          types.Address
	codeAddr      types.Address
	caller        *ContractNode
	enterMode     vm.OpCode
	dcfgNodeEntry *DcfgNode
	calls         []*callRef

	dcfgNodeCursor *DcfgNode
}

func NewContractNode(
	callId types.CallId,
	addr types.Address, codeAddr types.Address, caller *ContractNode, enterMode vm.OpCode,
) *ContractNode {
	return &ContractNode{
		callId:        callId,
		addr:          addr,
		codeAddr:      codeAddr,
		caller:        caller,
		enterMode:     enterMode,
		dcfgNodeEntry: nil,
		calls:         []*callRef{},

		dcfgNodeCursor: nil,
	}
}

func (contractNode *ContractNode) GetAddress() types.Address {
	return contractNode.addr
}

func (contractNode *ContractNode) GetCallId() types.CallId {
	return contractNode.callId
}

func (contractNode *ContractNode) GetEntry() *DcfgNode {
	return contractNode.dcfgNodeEntry
}

func (contractNode *ContractNode) GetCalls() []*ContractNode {
	calls := []*ContractNode{}
	for _, ref := range contractNode.calls {
		calls = append(calls, ref.contractNode)
	}
	return calls
}
