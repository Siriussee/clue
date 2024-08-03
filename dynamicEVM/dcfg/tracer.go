package dcfg

import (
	"fmt"
	"math/big"
	"time"

	"github.com/anon/execution-property-graph/dynamicEVM/calltrace"
	"github.com/anon/execution-property-graph/dynamicEVM/dataflow"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

type jumpi struct {
	id       string
	conditon *dataflow.DataSource
}

type jumpis struct {
	jumpis []*jumpi
}

type dataflowTracer interface {
	ReadStack(i int) *dataflow.DataSource
}

type DcfgTracer struct {
	callIdTracer   calltrace.CallIdTracer
	dataflowTracer dataflowTracer

	dcfg       *DCFG
	jumpiStack []*jumpis
	// jumpis map[string]*dataflow.DataSource

	isPreviousPush int
}

func NewDcfgTracer(
	callIdTracer calltrace.CallIdTracer,
) *DcfgTracer {
	return &DcfgTracer{
		callIdTracer:   callIdTracer,
		dcfg:           newDCFG(),
		jumpiStack:     []*jumpis{},
		isPreviousPush: -1,
	}
}

func (tracer *DcfgTracer) SetDataflowTracer(dataflowTracer dataflowTracer) {
	tracer.dataflowTracer = dataflowTracer
}

func (tracer *DcfgTracer) CaptureTxStart(gasLimit uint64) {}

func (tracer *DcfgTracer) CaptureTxEnd(restGas uint64) {}

func (tracer *DcfgTracer) CaptureStart(
	env types.EVM,
	from types.Address,
	to types.Address,
	create bool,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	tracer.jumpiStack = append(
		tracer.jumpiStack,
		&jumpis{
			jumpis: []*jumpi{},
		},
	)
	tracer.dcfg.Start(tracer.callIdTracer.GetCurrentCallId(), from, to, create)
}

func (tracer *DcfgTracer) CaptureState(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	rData []byte,
	depth int,
	err error,
) {
	// if the previous opcode is push
	if tracer.isPreviousPush != -1 {
		length := tracer.isPreviousPush - int(vm.PUSH1) + 1
		instructions := make([]byte, length)
		stack := scope.Stack.Data()
		top := stack[len(stack)-1].Bytes()
		if length < len(top) {
			panic("dcfgTracer: push length too long")
		}
		if len(top) > 0 {
			copy(instructions[length-len(top):], top)
		}
		tracer.dcfg.appendInstructions(pc-uint64(length), instructions)
		tracer.isPreviousPush = -1
	}

	var instructions []byte
	instructions = append(instructions, byte(op))

	if op.IsPush() {
		tracer.isPreviousPush = int(op)
	}

	// capture jumpi condition
	if op == vm.JUMPI {
		condition := tracer.dataflowTracer.ReadStack(1)
		currentJumpis := tracer.jumpiStack[len(tracer.jumpiStack)-1]
		currentJumpis.jumpis = append(
			currentJumpis.jumpis,
			&jumpi{
				id:       tracer.GetCurrentDcfgNodeId().String(),
				conditon: condition,
			},
		)
		// tracer.jumpis[tracer.GetCurrentDcfgNodeId().String()] = conditon
	}

	tracer.dcfg.addInstructions(tracer.callIdTracer.GetCurrentCallCount(), pc, instructions, scope)
}

func (tracer *DcfgTracer) CaptureEnter(
	typ vm.OpCode,
	from types.Address,
	to types.Address,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	tracer.jumpiStack = append(
		tracer.jumpiStack,
		&jumpis{
			jumpis: []*jumpi{},
		},
	)

	err := tracer.dcfg.Enter(tracer.callIdTracer.GetCurrentCallId(), from, to, typ)
	if err != nil {
		panic(fmt.Errorf("dcfg CaptureEnter error: %w", err))
	}
}

func (tracer *DcfgTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
	childJumpis := tracer.jumpiStack[len(tracer.jumpiStack)-1]
	tracer.jumpiStack = tracer.jumpiStack[:len(tracer.jumpiStack)-1]
	jumpis := tracer.jumpiStack[len(tracer.jumpiStack)-1]
	if err == nil {
		jumpis.jumpis = append(jumpis.jumpis, childJumpis.jumpis...)
	}

	tracer.dcfg.Exit()
}

func (tracer *DcfgTracer) CaptureFault(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	depth int,
	err error,
) {

}

func (tracer *DcfgTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {

}

func (tracer *DcfgTracer) GetCurrentCallId() types.CallId {
	contractNodeCursor := tracer.dcfg.contractNodeStack[len(tracer.dcfg.contractNodeStack)-1]

	return types.NewCallId(contractNodeCursor.callId)
}

func (tracer *DcfgTracer) GetCurrentDcfgNodeId() types.DcfgId {
	contractNodeCursor := tracer.dcfg.contractNodeStack[len(tracer.dcfg.contractNodeStack)-1]

	if contractNodeCursor.dcfgNodeCursor == nil {
		return types.NewDcfgId(contractNodeCursor.callId, 0, tracer.callIdTracer.GetCurrentCallCount())
	}

	return types.NewDcfgId(contractNodeCursor.callId, contractNodeCursor.dcfgNodeCursor.id, contractNodeCursor.dcfgNodeCursor.callCount)
}

func (tracer *DcfgTracer) GetDcfg() *DCFG {
	return tracer.dcfg
}

func (tracer *DcfgTracer) GetJumpiDataSource() map[string]*dataflow.DataSource {
	r := make(map[string]*dataflow.DataSource)
	for _, jumpi := range tracer.jumpiStack[0].jumpis {
		r[jumpi.id] = jumpi.conditon
	}

	return r
	// return tracer.jumpis
}
