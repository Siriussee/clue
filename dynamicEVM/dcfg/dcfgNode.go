package dcfg

type DcfgNode struct {
	contractNode *ContractNode
	basicBlock   *BasicBlock
	id           int
	callCount    int
	calls        []*callRef
	parent       *DcfgNode
	child        *DcfgNode
	opcode       int
	condition    bool

	pcCursor uint64
}

func (dcfgNode *DcfgNode) GetId() int {
	return dcfgNode.id
}

func (dcfgNode *DcfgNode) GetCallCount() int {
	return dcfgNode.callCount
}

func (dcfgNode *DcfgNode) GetPc() uint64 {
	return dcfgNode.basicBlock.pc
}

func (dcfgNode *DcfgNode) GetOpcode() int {
	return dcfgNode.opcode
}

func (dcfgNode *DcfgNode) GetCondition() bool {
	return dcfgNode.condition
}

func (dcfgNode *DcfgNode) GetChild() *DcfgNode {
	return dcfgNode.child
}

type callRef struct {
	id uint64
	pc uint64
	// from
	dcfgNode *DcfgNode
	// to
	contractNode *ContractNode
}
