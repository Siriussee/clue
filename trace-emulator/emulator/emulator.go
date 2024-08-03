package emulator

import (
	"errors"
	"fmt"

	"github.com/anon/execution-property-graph/dynamicEVM/tracer"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

type TraceEmulator struct {
	tracer tracer.Tracer
}

func NewTraceEmulator(tracer tracer.Tracer) *TraceEmulator {
	return &TraceEmulator{
		tracer: tracer,
	}
}

func (e *TraceEmulator) ExecuteTrace(
	trace *ExecutionResultBSON,
	callFrame *CallFrameBSON,
	header *types.Header,
) error {
	txFrom, txTo, txValue, txGas, txGasUsed, txInput, txOutput := callFrame.decode()

	create := false
	if callFrame.Type == "CREATE" {
		create = true
	}

	// TODO: fix env
	e.tracer.CaptureStart(nil, txFrom, txTo, create, txInput, txGas, txValue)

	var callErr error
	if callFrame.Error != "" {
		callErr = errors.New(callFrame.Error)
	}
	defer e.tracer.CaptureEnd(txOutput, txGasUsed, 0, callErr)

	callStack := []*callCursorBSON{
		{
			frame: callFrame,
			addr:  txTo,
		},
	}
	newCall := false

	for i, sr := range trace.StructLogs {
		// set cursor
		cursor := callStack[len(callStack)-1]

		// a new call should start with pc == 0
		// if not, then it is supposed to be a call to EOA
		if newCall && sr.Pc != 0 {
			_, _, _, _, gasUsed, _, output := cursor.frame.decode()

			var callErr error
			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}
			e.tracer.CaptureExit(output, gasUsed, callErr)

			// Only pop call stack if it is inter call
			if len(callStack) > 1 {
				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1
				cursor.rData = output
			}
		}
		newCall = false

		op := vm.StringToOp(sr.Op)
		scope := types.ScopeContext{
			Memory:   cpMemory(sr.Memory),
			Stack:    cpStack(sr.Stack),
			Contract: newContract(cursor.addr),
		}

		var stateErr error = nil
		if sr.Error {
			// erigon has some error for the structLog error
			// use the call frame error
			if cursor.frame.Error == "" {
				fmt.Println("WARNING: missing call error, emulation may be incorrect")
				// panic("missing call error")
			} else {
				stateErr = errors.New(cursor.frame.Error)
			}
		} else if cursor.frame.Error != "" {
			j := i + 1
			if j < len(trace.StructLogs) && trace.StructLogs[j].Depth < sr.Depth {
				stateErr = errors.New(cursor.frame.Error)
			}
		}

		e.tracer.CaptureState(sr.Pc, op, sr.Gas, sr.GasCost, scope, cursor.rData, sr.Depth, stateErr)

		if stateErr != nil {
			if len(callStack) > 1 {
				_, _, _, _, gasUsed, _, output := cursor.frame.decode()

				e.tracer.CaptureExit(output, gasUsed, stateErr)

				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1

				cursor.rData = output
			}
			continue
		}

		switch op {
		case vm.CALL, vm.CALLCODE, vm.DELEGATECALL, vm.STATICCALL, vm.CREATE, vm.CREATE2:
			// skip if calling into a precompiled contract
			toAddress := types.BytesToAddress((sr.Stack)[len(sr.Stack)-2][:])
			if op != vm.CREATE && op != vm.CREATE2 {
				if types.IsPrecompile(toAddress, header) {
					// For vultxtrace, the call frame does not include precompile call, disable the following code
					// For sereumtrace, the call frame includes precompile call, enable the following code
					cursor.index += 1
					continue
				}
			}

			var addr types.Address
			if op == vm.CALLCODE || op == vm.DELEGATECALL {
				addr = cursor.addr
			} else {
				addr = types.BytesToAddress(cursor.frame.Calls[cursor.index].To)
			}
			callStack = append(callStack, &callCursorBSON{
				frame: &cursor.frame.Calls[cursor.index],
				addr:  addr,
			})
			cursor = callStack[len(callStack)-1]

			if op.String() != cursor.frame.Type {
				return fmt.Errorf("invalid opcode expect %s, got %s", op.String(), cursor.frame.Type)
			}

			from, to, value, gas, _, input, _ := cursor.frame.decode()

			e.tracer.CaptureEnter(op, from, to, input, gas, value)

			// note that when the callee is not a smart contract
			// we would miss CaptureExit here
			newCall = true

		case vm.SELFDESTRUCT:
			callStack = append(callStack, &callCursorBSON{
				frame: &cursor.frame.Calls[cursor.index],
			})
			cursor = callStack[len(callStack)-1]

			from, to, value, gas, gasUsed, input, output := cursor.frame.decode()
			e.tracer.CaptureEnter(op, from, to, input, uint64(gas), value)

			var callErr error
			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}

			e.tracer.CaptureExit(output, gasUsed, callErr)

			callStack = callStack[:len(callStack)-1]
			cursor = callStack[len(callStack)-1]
			cursor.index += 1

			// SELFDESTRUCT stops the current call
			_, _, _, _, gasUsed, _, output = cursor.frame.decode()

			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}

			// Only pop call stack if it is inter call
			if len(callStack) > 1 {
				e.tracer.CaptureExit(output, gasUsed, callErr)

				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1

				cursor.rData = output
			}

		case vm.STOP, vm.RETURN, vm.REVERT:
			_, _, _, _, gasUsed, _, output := cursor.frame.decode()

			var callErr error
			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}

			// Only pop call stack if it is inter call
			if len(callStack) > 1 {
				e.tracer.CaptureExit(output, gasUsed, callErr)

				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1

				cursor.rData = output
			}
		}
	}

	return nil
}

func (e *TraceEmulator) ExecuteTraceRaw(
	trace *ExecutionResult,
	callFrame *CallFrame,
	header *types.Header,
) error {
	txFrom, txTo, txValue, txGas, txGasUsed, txInput, txOutput, err := callFrame.decode()
	if err != nil {
		return fmt.Errorf("decode frame error: %w", err)
	}

	create := false
	if callFrame.Type == "CREATE" {
		create = true
	}

	// TODO: fix env
	e.tracer.CaptureStart(nil, txFrom, txTo, create, txInput, txGas, txValue)

	var callErr error
	if callFrame.Error != "" {
		callErr = errors.New(callFrame.Error)
	}
	defer e.tracer.CaptureEnd(txOutput, txGasUsed, 0, callErr)

	callStack := []*callCursor{
		{
			frame: callFrame,
			addr:  txTo,
		},
	}
	newCall := false

	for i, sr := range trace.StructLogs {
		// set cursor
		cursor := callStack[len(callStack)-1]

		// a new call should start with pc == 0
		// if not, then it is supposed to be a call to EOA
		if newCall && sr.Pc != 0 {
			_, _, _, _, gasUsed, _, output, err := cursor.frame.decode()
			if err != nil {
				return fmt.Errorf("decode frame error: %w", err)
			}
			var callErr error
			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}
			e.tracer.CaptureExit(output, gasUsed, callErr)

			// Only pop call stack if it is inter call
			if len(callStack) > 1 {
				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1
				cursor.rData = output
			}
		}
		newCall = false

		op := vm.StringToOp(sr.Op)
		scope := types.ScopeContext{
			Memory:   newMemory(sr.Memory),
			Stack:    newStack(sr.Stack),
			Contract: newContract(cursor.addr),
		}

		var stateErr error = nil
		if sr.Error != nil {
			// erigon has some error for the structLog error
			// use the call frame error
			if cursor.frame.Error == "" {
				panic("missing call error")
			}
			stateErr = errors.New(cursor.frame.Error)
		} else if cursor.frame.Error != "" {
			j := i + 1
			if j < len(trace.StructLogs) && trace.StructLogs[j].Depth < sr.Depth {
				stateErr = errors.New(cursor.frame.Error)
			}
		}

		e.tracer.CaptureState(sr.Pc, op, sr.Gas, sr.GasCost, scope, cursor.rData, sr.Depth, stateErr)

		if stateErr != nil {
			if len(callStack) > 1 {
				_, _, _, _, gasUsed, _, output, err := cursor.frame.decode()
				if err != nil {
					return fmt.Errorf("decode frame error: %w", err)
				}

				e.tracer.CaptureExit(output, gasUsed, stateErr)

				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1

				cursor.rData = output
			}
			continue
		}

		switch op {
		case vm.CALL, vm.CALLCODE, vm.DELEGATECALL, vm.STATICCALL, vm.CREATE, vm.CREATE2:
			// skip if calling into a precompiled contract
			toAddress := types.HexToAddress((*sr.Stack)[len(*sr.Stack)-2])
			if op != vm.CREATE && op != vm.CREATE2 {
				if types.IsPrecompile(toAddress, header) {
					continue
				}
			}

			var addr types.Address
			if op == vm.CALLCODE || op == vm.DELEGATECALL {
				addr = cursor.addr
			} else {
				addr = types.HexToAddress(cursor.frame.Calls[cursor.index].To)
			}
			callStack = append(callStack, &callCursor{
				frame: &cursor.frame.Calls[cursor.index],
				addr:  addr,
			})
			cursor = callStack[len(callStack)-1]

			if op.String() != cursor.frame.Type {
				return fmt.Errorf("invalid opcode expect %s, got %s", op.String(), cursor.frame.Type)
			}

			from, to, value, gas, _, input, _, err := cursor.frame.decode()
			if err != nil {
				return fmt.Errorf("decode frame error: %w", err)
			}
			e.tracer.CaptureEnter(op, from, to, input, gas, value)

			// note that when the callee is not a smart contract
			// we would miss CaptureExit here
			newCall = true

		case vm.SELFDESTRUCT:
			callStack = append(callStack, &callCursor{
				frame: &cursor.frame.Calls[cursor.index],
			})
			cursor = callStack[len(callStack)-1]

			from, to, value, gas, gasUsed, input, output, err := cursor.frame.decode()
			if err != nil {
				return fmt.Errorf("decode frame error: %w", err)
			}
			e.tracer.CaptureEnter(op, from, to, input, uint64(gas), value)

			var callErr error
			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}

			e.tracer.CaptureExit(output, gasUsed, callErr)

			callStack = callStack[:len(callStack)-1]
			cursor = callStack[len(callStack)-1]
			cursor.index += 1

			// SELFDESTRUCT stops the current call
			_, _, _, _, gasUsed, _, output, err = cursor.frame.decode()
			if err != nil {
				return fmt.Errorf("decode frame error: %w", err)
			}

			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}

			// Only pop call stack if it is inter call
			if len(callStack) > 1 {
				e.tracer.CaptureExit(output, gasUsed, callErr)

				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1

				cursor.rData = output
			}

		case vm.STOP, vm.RETURN, vm.REVERT:
			_, _, _, _, gasUsed, _, output, err := cursor.frame.decode()
			if err != nil {
				return fmt.Errorf("decode frame error: %w", err)
			}

			var callErr error
			if cursor.frame.Error != "" {
				callErr = errors.New(cursor.frame.Error)
			}

			// Only pop call stack if it is inter call
			if len(callStack) > 1 {
				e.tracer.CaptureExit(output, gasUsed, callErr)

				callStack = callStack[:len(callStack)-1]
				cursor = callStack[len(callStack)-1]
				cursor.index += 1

				cursor.rData = output
			}
		}
	}

	return nil
}
