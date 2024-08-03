package tracer

import (
	"math/big"
	"time"

	"github.com/anon/execution-property-graph/dynamicEVM/calltrace"
	"github.com/anon/execution-property-graph/dynamicEVM/dataflow"
	"github.com/anon/execution-property-graph/dynamicEVM/dcfg"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
	"github.com/holiman/uint256"
)

var ERC20_TRANSFER_TOPIC, _ = uint256.FromHex("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

var big0 = big.NewInt(0)

type CallTracer struct {
	calltrace.CallTraces

	currentCallTrace *calltrace.CallTrace
	callBoundaries   []int
	callStack        []vm.OpCode

	lastDcfgId types.DcfgId

	callIdTracer   calltrace.CallIdTracer
	dcfgTracer     *dcfg.DcfgTracer
	dataflowTracer *dataflow.DataflowTracer
}

func NewCallTracer(
	callIdTracer calltrace.CallIdTracer,
	dcfgTracer *dcfg.DcfgTracer,
	dataflowTracer *dataflow.DataflowTracer,
) *CallTracer {
	return &CallTracer{
		lastDcfgId:     types.NewDcfgId(types.NewCallId([]int{}), 0, 0),
		callIdTracer:   callIdTracer,
		dcfgTracer:     dcfgTracer,
		dataflowTracer: dataflowTracer,
	}
}

func (tracer *CallTracer) recordEtherFlow(from, to types.Address, amount *big.Int, amountTracker *dataflow.DataSource) {
	if amount.Cmp(big0) > 0 {
		dcfgId := tracer.lastDcfgId
		currentIdx := len(tracer.currentCallTrace.Flows)
		flow := calltrace.NewEtherFlow(from, to, amount, currentIdx, dcfgId, amountTracker)
		tracer.currentCallTrace.Parent.Flows = append(tracer.currentCallTrace.Parent.Flows, flow)
	}
}

func (tracer *CallTracer) recordERC20Flow(from, to, asset types.Address, amount *big.Int, amountTracker *dataflow.DataSource) {
	if amount.Cmp(big0) > 0 {
		dcfgId := tracer.dcfgTracer.GetCurrentDcfgNodeId()
		currentIdx := len(tracer.currentCallTrace.Flows)
		flow := calltrace.NewERC20Flow(from, to, asset, amount, currentIdx, dcfgId, amountTracker)
		tracer.currentCallTrace.Flows = append(tracer.currentCallTrace.Flows, flow)
	}
}

func (tracer *CallTracer) recordCall(typ vm.OpCode, from, to types.Address, amount *big.Int, amountTracker *dataflow.DataSource) {
	trace := &calltrace.CallTrace{
		ID:     tracer.callIdTracer.GetCurrentCallId(),
		From:   from,
		To:     to,
		Type:   typ,
		Parent: tracer.currentCallTrace,
	}
	if tracer.currentCallTrace == nil {
		trace.Parent = trace
	}
	tracer.currentCallTrace = trace
	tracer.callBoundaries = append(tracer.callBoundaries, len(tracer.CallTraces))
	tracer.CallTraces = append(tracer.CallTraces, trace)
	tracer.recordEtherFlow(from, to, amount, amountTracker)
}

func (tracer *CallTracer) returnCall() {
	if len(tracer.callBoundaries) > 0 {
		tracer.callBoundaries = tracer.callBoundaries[:len(tracer.callBoundaries)-1]
		if len(tracer.callBoundaries) > 0 {
			tracer.currentCallTrace = tracer.CallTraces[tracer.callBoundaries[len(tracer.callBoundaries)-1]]
		} else {
			tracer.currentCallTrace = nil
		}
	}
}

func (tracer *CallTracer) rewindCall() {
	l := len(tracer.callBoundaries)
	if l > 0 {
		tracer.CallTraces = tracer.CallTraces[:tracer.callBoundaries[l-1]]
		tracer.callBoundaries = tracer.callBoundaries[:l-1]
		if len(tracer.callBoundaries) > 0 {
			tracer.currentCallTrace = tracer.CallTraces[len(tracer.CallTraces)-1]
		} else {
			tracer.currentCallTrace = nil
		}
	}
}

// Tracer part

func (tracer *CallTracer) CaptureTxStart(gasLimit uint64) {}

func (tracer *CallTracer) CaptureTxEnd(restGas uint64) {}

func (tracer *CallTracer) CaptureStart(
	env types.EVM,
	from types.Address,
	to types.Address,
	create bool,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	var typ vm.OpCode
	if create {
		typ = vm.CREATE
	} else {
		typ = vm.CALL
	}
	tracer.callStack = append(tracer.callStack, typ)
	tracer.recordCall(typ, from, to, value, nil)
}

func (tracer *CallTracer) CaptureState(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	rData []byte,
	depth int,
	err error,
) {
	if op == vm.LOG3 {
		data := scope.Stack.Data()
		topic := data[len(data)-3]
		if topic.Eq(ERC20_TRANSFER_TOPIC) {
			from := types.BytesToAddress(data[len(data)-4].Bytes())
			to := types.BytesToAddress(data[len(data)-5].Bytes())
			asset := scope.Contract.Address
			mStart := int64(data[len(data)-1].Uint64())
			mSize := int64(data[len(data)-2].Uint64())
			amount := new(big.Int).SetBytes(scope.Memory.Data()[mStart : mStart+mSize])

			var amountTracker *dataflow.DataSource
			if tracer.dataflowTracer != nil {
				amountTracker = tracer.dataflowTracer.ReadMemory(uint64(mStart), uint64(mSize))
			}
			tracer.recordERC20Flow(from, to, asset, amount, amountTracker)
		}
	}
	tracer.lastDcfgId = tracer.dcfgTracer.GetCurrentDcfgNodeId()
}

func (tracer *CallTracer) CaptureEnter(
	typ vm.OpCode,
	from types.Address,
	to types.Address,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	var amountTracker *dataflow.DataSource

	tracer.callStack = append(tracer.callStack, typ)

	if tracer.dataflowTracer != nil {
		amountTracker = tracer.dataflowTracer.ReadCallValue()
	}

	tracer.recordCall(typ, from, to, value, amountTracker)
}

func (tracer *CallTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
	if err == nil {
		tracer.returnCall()
	} else {
		tracer.rewindCall()
	}
	tracer.callStack = tracer.callStack[:len(tracer.callStack)-1]
}

func (tracer *CallTracer) CaptureFault(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	depth int,
	err error,
) {
}

func (tracer *CallTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
	tracer.returnCall() // Unnecessary, only to clear callBoundaries
	tracer.callStack = tracer.callStack[:len(tracer.callStack)-1]
}
