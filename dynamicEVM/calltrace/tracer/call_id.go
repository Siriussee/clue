package tracer

import (
	"math/big"
	"time"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

type CallIdTracer struct {
	callFrameIds []int
}

func NewCallIdTracer() *CallIdTracer {
	return &CallIdTracer{}
}

func (tracer *CallIdTracer) GetCurrentCallId() types.CallId {
	return types.NewCallId(tracer.callFrameIds[:len(tracer.callFrameIds)-1])
}

func (tracer *CallIdTracer) GetCurrentCallCount() int {
	return tracer.callFrameIds[len(tracer.callFrameIds)-1]
}

func (tracer *CallIdTracer) CaptureTxStart(gasLimit uint64) {}

func (tracer *CallIdTracer) CaptureTxEnd(restGas uint64) {}

func (tracer *CallIdTracer) CaptureStart(
	env types.EVM,
	from types.Address,
	to types.Address,
	create bool,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	tracer.callFrameIds = append(tracer.callFrameIds, 0)
}

func (tracer *CallIdTracer) CaptureState(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	rData []byte,
	depth int,
	err error,
) {
}

func (tracer *CallIdTracer) CaptureEnter(
	typ vm.OpCode,
	from types.Address,
	to types.Address,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	tracer.callFrameIds = append(tracer.callFrameIds, 0)
}

func (tracer *CallIdTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
	tracer.callFrameIds = tracer.callFrameIds[:len(tracer.callFrameIds)-1]
	tracer.callFrameIds[len(tracer.callFrameIds)-1] += 1
}

func (tracer *CallIdTracer) CaptureFault(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	depth int,
	err error,
) {
}

func (tracer *CallIdTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
}
