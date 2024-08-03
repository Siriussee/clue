package tracer

import (
	"math/big"
	"time"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

type Tracer interface {
	// Transaction level
	CaptureTxStart(gasLimit uint64)
	CaptureTxEnd(restGas uint64)
	// Top call frame
	CaptureStart(env types.EVM, from types.Address, to types.Address, create bool, input []byte, gas uint64, value *big.Int)
	CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error)
	// Rest of call frames
	CaptureEnter(typ vm.OpCode, from types.Address, to types.Address, input []byte, gas uint64, value *big.Int)
	CaptureExit(output []byte, gasUsed uint64, err error)
	// Opcode level
	CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope types.ScopeContext, rData []byte, depth int, err error)
	CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope types.ScopeContext, depth int, err error)
}

type evmTracer struct {
	subTracers []Tracer
}

func NewEvmTracer() *evmTracer {
	return &evmTracer{}
}

func (tracer *evmTracer) Register(st Tracer) int {
	tracer.subTracers = append(tracer.subTracers, st)
	return len(tracer.subTracers) - 1
}

func (tracer *evmTracer) Unregister(id int) {
	if id < len(tracer.subTracers) {
		tracer.subTracers = append(tracer.subTracers[:id], tracer.subTracers[id+1:]...)
	}
}

func (tracer *evmTracer) CaptureTxStart(gasLimit uint64) {}

func (tracer *evmTracer) CaptureTxEnd(restGas uint64) {}

func (tracer *evmTracer) CaptureStart(
	env types.EVM,
	from types.Address,
	to types.Address,
	create bool,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	for _, st := range tracer.subTracers {
		st.CaptureStart(env, from, to, create, input, gas, value)
	}
}

func (tracer *evmTracer) CaptureState(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	rData []byte,
	depth int,
	err error,
) {
	for _, st := range tracer.subTracers {
		st.CaptureState(pc, op, gas, cost, scope, rData, depth, err)
	}
}

func (tracer *evmTracer) CaptureEnter(
	typ vm.OpCode,
	from types.Address,
	to types.Address,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	for _, st := range tracer.subTracers {
		st.CaptureEnter(typ, from, to, input, gas, value)
	}
}

func (tracer *evmTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
	for _, st := range tracer.subTracers {
		st.CaptureExit(output, gasUsed, err)
	}
}

func (tracer *evmTracer) CaptureFault(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	depth int,
	err error,
) {
	for _, st := range tracer.subTracers {
		st.CaptureFault(pc, op, gas, cost, scope, depth, err)
	}
}

func (tracer *evmTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
	for _, st := range tracer.subTracers {
		st.CaptureEnd(output, gasUsed, t, err)
	}
}
