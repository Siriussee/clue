package graph

import (
	calltraceTracer "github.com/anon/execution-property-graph/dynamicEVM/calltrace/tracer"
	"github.com/anon/execution-property-graph/dynamicEVM/dataflow"
	"github.com/anon/execution-property-graph/dynamicEVM/dcfg"
	"github.com/anon/execution-property-graph/dynamicEVM/tracer"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	emu "github.com/anon/execution-property-graph/trace-emulator/emulator"
)

type TraceResult struct {
	*calltraceTracer.CallTracer
	*dcfg.DcfgTracer
	*dataflow.DataflowTracer
}

func TraceOnEmulator(
	txHash types.Hash,
	trace *emu.ExecutionResult,
	callFrame *emu.CallFrame,
	header *types.Header,
) (*TraceResult, error) {
	txBytes := types.HexToBytes32(txHash.Hex())
	traceBSON, structLogBSON, err := trace.ToBSON(txBytes)
	if err != nil {
		return nil, err
	}
	traceBSON.StructLogs = structLogBSON
	callFrameBSON, err := callFrame.ToBSON()
	if err != nil {
		return nil, err
	}

	return TraceBSONOnEmulator(&traceBSON, &callFrameBSON, header)
}

func TraceBSONOnEmulator(
	traceBSON *emu.ExecutionResultBSON,
	callFrameBSON *emu.CallFrameBSON,
	header *types.Header,
) (*TraceResult, error) {
	tracer := tracer.NewEvmTracer()

	callIdTracer := calltraceTracer.NewCallIdTracer()
	dcfgTracer := dcfg.NewDcfgTracer(callIdTracer)
	dataflowTracer := dataflow.NewDataFlowTracer()
	dataflowTracer.SetDcfgTracer(dcfgTracer)
	dcfgTracer.SetDataflowTracer(dataflowTracer)
	callTracer := calltraceTracer.NewCallTracer(callIdTracer, dcfgTracer, dataflowTracer)
	tracer.Register(callIdTracer)
	tracer.Register(dcfgTracer)
	tracer.Register(callTracer)
	tracer.Register(dataflowTracer)

	emulator := emu.NewTraceEmulator(tracer)
	err := emulator.ExecuteTrace(traceBSON, callFrameBSON, header)
	if err != nil {
		return nil, err
	}

	return &TraceResult{
		CallTracer:     callTracer,
		DcfgTracer:     dcfgTracer,
		DataflowTracer: dataflowTracer,
	}, nil
}
