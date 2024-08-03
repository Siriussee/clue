package dataflow

import (
	"math/big"
	"time"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

type dcfgTracer interface {
	GetCurrentDcfgNodeId() types.DcfgId
	GetCurrentCallId() types.CallId
}

type DataflowTracer struct {
	dcfgTracer dcfgTracer

	ctxStack []*context
	stg      *storage
	bal      *balance
}

func NewDataFlowTracer() *DataflowTracer {
	return &DataflowTracer{
		ctxStack: []*context{},
		stg:      newStorage(),
		bal:      newBalance(),
	}
}

func (tracer *DataflowTracer) SetDcfgTracer(dcfgTracer dcfgTracer) {
	tracer.dcfgTracer = dcfgTracer
}

func (tracer *DataflowTracer) CaptureTxStart(gasLimit uint64) {}

func (tracer *DataflowTracer) CaptureTxEnd(restGas uint64) {}

func (tracer *DataflowTracer) CaptureStart(
	env types.EVM,
	from types.Address,
	to types.Address,
	create bool,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	callId := tracer.dcfgTracer.GetCurrentCallId()
	ctx := &context{
		callId:          callId,
		addr:            to,
		codeAddr:        to,
		caller:          from,
		mem:             newMemory(),
		sta:             newStack(),
		callValue:       newCallValue(callId, newNilWord()),
		childCallData:   []*DataSource{},
		callData:        newCallData(callId, make([]*DataSource, len(input))),
		returnData:      []*DataSource{},
		childReturnData: []*DataSource{},
	}

	if value.Cmp(new(big.Int)) > 0 {
		hid := tracer.bal.writeBalance(to, ctx.callValue, callId)
		ctx.balanceWrites = append(
			ctx.balanceWrites,
			&BalanceLocation{
				addr: to,
				hid:  hid,
			},
		)
	}

	tracer.ctxStack = append(tracer.ctxStack, ctx)
}

func (tracer *DataflowTracer) CaptureState(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	rData []byte,
	depth int,
	err error,
) {
	if depth != len(tracer.ctxStack) {
		panic("ctx size mismatch")
	}
	ctx := tracer.ctxStack[len(tracer.ctxStack)-1]
	// if len(ctx.sta.s) != len(scope.Stack.Data()) {
	// 	panic("stack size mismatch")
	// }

	exec(pc, op, ctx, tracer.stg, tracer.bal, scope, tracer.dcfgTracer)
}

func (tracer *DataflowTracer) CaptureEnter(
	typ vm.OpCode,
	from types.Address,
	to types.Address,
	input []byte,
	gas uint64,
	value *big.Int,
) {
	var (
		addr      types.Address
		caller    types.Address
		callValue word
	)

	preCtx := tracer.ctxStack[len(tracer.ctxStack)-1]

	if typ == vm.CALLCODE || typ == vm.DELEGATECALL {
		addr = from
	} else {
		addr = to
	}

	if typ == vm.DELEGATECALL {
		caller = preCtx.caller
		callValue = preCtx.callValue
	} else {
		caller = from
		callValue = preCtx.childCallValue
	}

	callId := tracer.dcfgTracer.GetCurrentCallId()

	ctx := &context{
		callId:          callId,
		addr:            addr,
		codeAddr:        to,
		caller:          caller,
		mem:             newMemory(),
		sta:             newStack(),
		callValue:       callValue,
		childCallData:   []*DataSource{},
		callData:        newCallData(callId, preCtx.childCallData),
		returnData:      []*DataSource{},
		childReturnData: []*DataSource{},
	}

	if typ == vm.CALL || typ == vm.CREATE || typ == vm.CREATE2 || typ == vm.SELFDESTRUCT {
		if value.Cmp(new(big.Int)) > 0 {
			hid := tracer.bal.writeBalance(to, callValue, callId)
			ctx.balanceWrites = append(
				ctx.balanceWrites,
				&BalanceLocation{
					addr: to,
					hid:  hid,
				},
			)
		}
	}

	tracer.ctxStack = append(tracer.ctxStack, ctx)
}

func (tracer *DataflowTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
	childCtx := tracer.ctxStack[len(tracer.ctxStack)-1]
	tracer.ctxStack = tracer.ctxStack[:len(tracer.ctxStack)-1]
	ctx := tracer.ctxStack[len(tracer.ctxStack)-1]
	ctx.childReturnData = make([]*DataSource, len(childCtx.returnData))
	copy(ctx.childReturnData, childCtx.returnData)

	ctx.mem.mstoreChunk(ctx.retOffset, ctx.retLength, childCtx.returnData)
	ctx.retOffset = 0
	ctx.retLength = 0

	ctx.sta.push(newExecResult(childCtx.callId))

	if err == nil {
		ctx.storageWrites = append(ctx.storageWrites, childCtx.storageWrites...)
		ctx.balanceWrites = append(ctx.balanceWrites, childCtx.balanceWrites...)
	} else {
		for _, w := range childCtx.storageWrites {
			tracer.stg.revert(w.addr, w.slot, w.hid)
		}
		for _, w := range childCtx.balanceWrites {
			tracer.bal.revert(w.addr, w.hid)
		}
	}
}

func (tracer *DataflowTracer) CaptureFault(
	pc uint64,
	op vm.OpCode,
	gas, cost uint64,
	scope types.ScopeContext,
	depth int,
	err error,
) {

}

func (tracer *DataflowTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {

}

func (tracer *DataflowTracer) ReadMemory(start uint64, size uint64) *DataSource {
	ctx := tracer.ctxStack[len(tracer.ctxStack)-1]
	sources := ctx.mem.mloadChunk(start, size)

	return mergeDataSources(sources...)
}

func (tracer *DataflowTracer) ReadStack(pos int) *DataSource {
	ctx := tracer.ctxStack[len(tracer.ctxStack)-1]

	return mergeDataSources(ctx.sta.s[len(ctx.sta.s)-1-pos][:]...)
}

func (tracer *DataflowTracer) ReadCallValue() *DataSource {
	ctx := tracer.ctxStack[len(tracer.ctxStack)-1]

	return mergeDataSources(ctx.childCallValue[:]...)
}

func (tracer *DataflowTracer) GetStorageHistory(loc *StorageLocation) []*StorageValue {
	return tracer.stg.s[loc.addr][loc.slot]
}

func (tracer *DataflowTracer) GetBalanceHistory(loc *BalanceLocation) []*BalanceValue {
	return tracer.bal.b[loc.addr]
}
