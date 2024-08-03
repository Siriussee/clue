package dcfg

import (
	"errors"
	"fmt"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

// Dynamic Control-flow Graph
type DCFG struct {
	basicBlocks         map[types.Address]map[uint64]*BasicBlock
	creationBasicBlocks map[types.Address]map[uint64]*BasicBlock
	root                *ContractNode

	contractNodeStack []*ContractNode
}

func (dcfg *DCFG) GetRoot() *ContractNode {
	return dcfg.root
}

func newDCFG() *DCFG {
	return &DCFG{
		basicBlocks:         make(map[types.Address]map[uint64]*BasicBlock),
		creationBasicBlocks: make(map[types.Address]map[uint64]*BasicBlock),
		root:                nil,

		contractNodeStack: []*ContractNode{},
	}
}

func (dcfg *DCFG) Start(callId types.CallId, from types.Address, to types.Address, create bool) {
	var enterMode vm.OpCode
	if create {
		enterMode = vm.CREATE
	} else {
		enterMode = vm.CALL
	}

	root := NewContractNode(callId, to, to, nil, enterMode)
	dcfg.root = root

	dcfg.contractNodeStack = append(dcfg.contractNodeStack, root)
}

func (dcfg *DCFG) Enter(callId types.CallId, from types.Address, to types.Address, enterMode vm.OpCode) error {
	codeAddr := to

	var addr types.Address
	switch enterMode {
	case vm.CALL, vm.STATICCALL, vm.CREATE, vm.CREATE2, vm.SELFDESTRUCT:
		addr = to
	case vm.DELEGATECALL, vm.CALLCODE:
		addr = from
	default:
		return errors.New("wrong enterMode")
	}

	contractNodeCursor := dcfg.contractNodeStack[len(dcfg.contractNodeStack)-1]
	contractNode := NewContractNode(callId, addr, codeAddr, dcfg.contractNodeStack[len(dcfg.contractNodeStack)-1], enterMode)

	ref := &callRef{
		id:           uint64(len(contractNodeCursor.dcfgNodeCursor.calls)),
		pc:           contractNodeCursor.dcfgNodeCursor.pcCursor,
		dcfgNode:     contractNodeCursor.dcfgNodeCursor,
		contractNode: contractNode,
	}

	contractNodeCursor.dcfgNodeCursor.calls = append(contractNodeCursor.dcfgNodeCursor.calls, ref)
	contractNodeCursor.calls = append(contractNodeCursor.calls, ref)

	dcfg.contractNodeStack = append(dcfg.contractNodeStack, contractNode)

	return nil
}

func (dcfg *DCFG) Exit() {
	dcfg.contractNodeStack = dcfg.contractNodeStack[:len(dcfg.contractNodeStack)-1]
}

func (dcfg *DCFG) appendInstructions(
	pc uint64,
	instructions []byte,
) error {
	contractNodeCursor := dcfg.contractNodeStack[len(dcfg.contractNodeStack)-1]

	return contractNodeCursor.dcfgNodeCursor.basicBlock.addInstructions(pc, instructions)
}

func (dcfg *DCFG) addInstructions(
	callCount int,
	pc uint64,
	instructions []byte,
	scope types.ScopeContext,
) error {
	contractNodeCursor := dcfg.contractNodeStack[len(dcfg.contractNodeStack)-1]

	if contractNodeCursor.dcfgNodeCursor == nil {
		contractNodeCursor.dcfgNodeCursor = &DcfgNode{
			contractNode: contractNodeCursor,
			basicBlock:   dcfg.GetBasicBlock(contractNodeCursor, pc),
			id:           0,
			callCount:    callCount,
			calls:        []*callRef{},
			parent:       nil,
			child:        nil,
			opcode:       -1,
			condition:    false,
		}
		contractNodeCursor.dcfgNodeEntry = contractNodeCursor.dcfgNodeCursor
	}

	// some basic blocks has JUMPDEST in the middle
	// to avoid duplications
	// here we separate
	if vm.OpCode(instructions[0]) == vm.JUMPDEST && contractNodeCursor.dcfgNodeCursor.basicBlock.pc != INVALIDPC && contractNodeCursor.dcfgNodeCursor.basicBlock.pc != pc {
		childDcfgNode := &DcfgNode{
			contractNode: contractNodeCursor,
			basicBlock:   dcfg.GetBasicBlock(contractNodeCursor, pc),
			id:           contractNodeCursor.dcfgNodeCursor.id + 1,
			callCount:    callCount,
			calls:        []*callRef{},
			parent:       contractNodeCursor.dcfgNodeCursor,
			child:        nil,
			opcode:       -1,
			condition:    false,
		}
		contractNodeCursor.dcfgNodeCursor.child = childDcfgNode
		contractNodeCursor.dcfgNodeCursor = childDcfgNode
	}

	// add instructions to block
	err := contractNodeCursor.dcfgNodeCursor.basicBlock.addInstructions(pc, instructions)
	if err != nil {
		return fmt.Errorf("addInstructions error: %w", err)
	}

	// update pcCursor when it's a call
	if _, exist := EnterOp[vm.OpCode(instructions[0])]; exist {
		contractNodeCursor.dcfgNodeCursor.pcCursor = pc
	}

	// handle with block end
	if _, exist := BasicBlockEnd[vm.OpCode(instructions[0])]; exist {
		if vm.OpCode(instructions[0]) == vm.JUMPI {
			stack := scope.Stack.Data()
			condition := !stack[len(stack)-2].IsZero()
			var destPc uint64
			if condition {
				destPc = stack[len(stack)-1].Uint64()
			} else {
				destPc = pc + 1
			}
			childDcfgNode := &DcfgNode{
				contractNode: contractNodeCursor,
				basicBlock:   dcfg.GetBasicBlock(contractNodeCursor, destPc),
				id:           contractNodeCursor.dcfgNodeCursor.id + 1,
				callCount:    callCount,
				calls:        []*callRef{},
				parent:       contractNodeCursor.dcfgNodeCursor,
				child:        nil,
				opcode:       int(vm.JUMPI),
				condition:    condition,
			}
			contractNodeCursor.dcfgNodeCursor.child = childDcfgNode
			contractNodeCursor.dcfgNodeCursor = childDcfgNode
		} else if vm.OpCode(instructions[0]) == vm.JUMP {
			stack := scope.Stack.Data()
			destPc := stack[len(stack)-1].Uint64()
			childDcfgNode := &DcfgNode{
				contractNode: contractNodeCursor,
				basicBlock:   dcfg.GetBasicBlock(contractNodeCursor, destPc),
				id:           contractNodeCursor.dcfgNodeCursor.id + 1,
				callCount:    callCount,
				calls:        []*callRef{},
				parent:       contractNodeCursor.dcfgNodeCursor,
				child:        nil,
				opcode:       int(vm.JUMP),
				condition:    true,
			}
			contractNodeCursor.dcfgNodeCursor.child = childDcfgNode
			contractNodeCursor.dcfgNodeCursor = childDcfgNode
		} else if vm.OpCode(instructions[0]) != vm.SELFDESTRUCT {
			contractNodeCursor.dcfgNodeCursor = nil
		}
	}

	return nil
}

func (dcfg *DCFG) GetBasicBlock(contractNodeCursor *ContractNode, pc uint64) *BasicBlock {
	if contractNodeCursor.enterMode == vm.CREATE || contractNodeCursor.enterMode == vm.CREATE2 {
		if _, exist := dcfg.creationBasicBlocks[contractNodeCursor.codeAddr]; !exist {
			dcfg.creationBasicBlocks[contractNodeCursor.codeAddr] = make(map[uint64]*BasicBlock)
		}
		if _, exist := dcfg.creationBasicBlocks[contractNodeCursor.codeAddr][pc]; !exist {
			dcfg.creationBasicBlocks[contractNodeCursor.codeAddr][pc] = NewBasicBlock(Creation, contractNodeCursor.codeAddr)
		}
		return dcfg.creationBasicBlocks[contractNodeCursor.codeAddr][pc]
	} else {
		if _, exist := dcfg.basicBlocks[contractNodeCursor.codeAddr]; !exist {
			dcfg.basicBlocks[contractNodeCursor.codeAddr] = make(map[uint64]*BasicBlock)
		}
		if _, exist := dcfg.basicBlocks[contractNodeCursor.codeAddr][pc]; !exist {
			dcfg.basicBlocks[contractNodeCursor.codeAddr][pc] = NewBasicBlock(Runtime, contractNodeCursor.codeAddr)
		}
		return dcfg.basicBlocks[contractNodeCursor.codeAddr][pc]
	}
}
