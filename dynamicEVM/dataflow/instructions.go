package dataflow

import (
	"fmt"
	"math/big"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

func exec(pc uint64, op vm.OpCode, ctx *context, stg *storage, bal *balance, scope types.ScopeContext, dcfg dcfgTracer) {
	switch {
	case op == vm.ORIGIN:
		ctx.sta.push(newWordWithNilAncestor(Origin, emptyLoc, 20))
	case op == vm.CALLER:
		ctx.sta.push(newWordWithNilAncestor(Caller, ctx.callId, 20))
	case uint64(op) >= uint64(vm.ADD) && uint64(op) <= uint64(vm.SMOD):
		ctx.sta.peekNPush(2)
	case uint64(op) >= uint64(vm.ADDMOD) && uint64(op) <= uint64(vm.MULMOD):
		ctx.sta.peekNPush(3)
	case uint64(op) >= uint64(vm.EXP) && uint64(op) <= uint64(vm.SIGNEXTEND):
		ctx.sta.peekNPush(2)
	case uint64(op) >= uint64(vm.LT) && uint64(op) <= uint64(vm.EQ):
		ctx.sta.peekNPush(2)
	case op == vm.ISZERO:
	case uint64(op) >= uint64(vm.AND) && uint64(op) <= uint64(vm.XOR):
		x := ctx.sta.s[len(ctx.sta.s)-1]
		y := ctx.sta.s[len(ctx.sta.s)-2]
		e := wordMerge(x, y)
		ctx.sta.pop(2)
		ctx.sta.push(e)
	case op == vm.NOT:
	case op == vm.BYTE:
		stack := scope.Stack.Data()
		e := newNilWord()
		if number, overflow := stack[len(stack)-1].Uint64WithOverflow(); !overflow {
			if number < 32 {
				e[31] = ctx.sta.s[len(ctx.sta.s)-1][number]
			}
		}
		ctx.sta.pop(2)
		ctx.sta.push(e)
	case uint64(op) >= uint64(vm.SHL) && uint64(op) <= uint64(vm.SAR):
		stack := scope.Stack.Data()
		shift := stack[len(stack)-1]
		w := newNilWord()
		if shift.LtUint64(256) {
			e := mergeDataSources(ctx.sta.s[len(ctx.sta.s)-2][:]...)
			for i := 0; i < 32; i++ {
				w[i] = e
			}
		}
		ctx.sta.pop(2)
		ctx.sta.push(w)
	case op == vm.KECCAK256:
		stack := scope.Stack.Data()
		offset := stack[len(stack)-1].Uint64()
		size := stack[len(stack)-2].Uint64()
		val := mergeDataSources(ctx.mem.mloadChunk(offset, size)...)
		w := newNilWord()
		for i := 0; i < 32; i++ {
			w[i] = val
		}

		ctx.sta.pop(2)
		ctx.sta.push(w)
	case op == vm.ADDRESS:
		ctx.sta.push(newWordWithNilAncestor(Address, ctx.callId, 20))
	case op == vm.BALANCE:
		stack := scope.Stack.Data()
		addr := types.BytesToAddress(stack[len(stack)-1].Bytes())

		ctx.sta.pop(1)
		ctx.sta.push(bal.readBalance(addr, ctx.callId))
	case op == vm.CALLVALUE:
		ctx.sta.push(ctx.callValue)
	case op == vm.CALLDATALOAD:
		stack := scope.Stack.Data()

		var data []*DataSource
		if offset, overflow := stack[len(stack)-1].Uint64WithOverflow(); !overflow {
			data = getData(ctx.callData, offset, 32)
		} else {
			data = make([]*DataSource, 32)
		}

		var w word
		copy(w[:], data)
		ctx.sta.pop(1)
		ctx.sta.push(w)
	case op == vm.CALLDATASIZE:
		ctx.sta.push(newWordWithNilAncestor(CallDataSize, ctx.callId, 32))
	case op == vm.CALLDATACOPY:
		stack := scope.Stack.Data()
		memOffset := stack[len(stack)-1]
		dataOffset := stack[len(stack)-2]
		length := stack[len(stack)-3]

		dataOffset64, overflow := dataOffset.Uint64WithOverflow()
		if overflow {
			dataOffset64 = 0xffffffffffffffff
		}
		// These values are checked for overflow during gas cost calculation
		memOffset64 := memOffset.Uint64()
		length64 := length.Uint64()

		data := getData(ctx.callData, dataOffset64, length64)
		ctx.mem.mstoreChunk(memOffset64, length64, data)

		ctx.sta.pop(3)
	case op == vm.CODESIZE:
		ctx.sta.push(newWordWithNilAncestor(CodeSize, NewAddressLocation(ctx.codeAddr), 32))
	case op == vm.CODECOPY:
		stack := scope.Stack.Data()
		memOffset := stack[len(stack)-1]
		codeOffset := stack[len(stack)-2]
		length := stack[len(stack)-3]

		uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
		if overflow {
			uint64CodeOffset = 0xffffffffffffffff
		}
		codeCopy := getData(
			newCode(ctx.codeAddr, uint64CodeOffset, length.Uint64()),
			uint64CodeOffset,
			length.Uint64(),
		)
		ctx.mem.mstoreChunk(memOffset.Uint64(), length.Uint64(), codeCopy)

		ctx.sta.pop(3)
	case op == vm.GASPRICE:
		ctx.sta.push(newWordWithNilAncestor(GasPrice, emptyLoc, 32))
	case op == vm.EXTCODESIZE:
		stack := scope.Stack.Data()
		addr := types.BytesToAddress(stack[len(stack)-1].Bytes())

		ctx.sta.pop(1)
		ctx.sta.push(newWordWithNilAncestor(ExtCodeSize, NewAddressLocation(addr), 32))
	case op == vm.EXTCODECOPY:
		stack := scope.Stack.Data()
		addr := stack[len(stack)-1].Bytes20()
		memOffset := stack[len(stack)-2]
		codeOffset := stack[len(stack)-3]
		length := stack[len(stack)-4]

		uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
		if overflow {
			uint64CodeOffset = 0xffffffffffffffff
		}

		codeCopy := getData(
			newCode(types.BytesToAddress(addr[:]), uint64CodeOffset, length.Uint64()),
			uint64CodeOffset,
			length.Uint64(),
		)
		ctx.mem.mstoreChunk(memOffset.Uint64(), length.Uint64(), codeCopy)

		ctx.sta.pop(4)
	case op == vm.RETURNDATASIZE:
		// ctx.sta.push(newWordWithNilAncestor(ReturnDataSize, emptyLoc, 32)) //TODO
		ctx.sta.push(newWordWithNilAncestor(ReturnDataSize, ctx.callId, 32))
	case op == vm.RETURNDATACOPY:
		stack := scope.Stack.Data()
		memOffset := stack[len(stack)-1]
		dataOffset := stack[len(stack)-2]
		length := stack[len(stack)-3]

		offset64, overflow := dataOffset.Uint64WithOverflow()
		if overflow {
			return
		}
		// we can reuse dataOffset now (aliasing it for clarity)
		var end = dataOffset
		end.Add(&dataOffset, &length)
		end64, overflow := end.Uint64WithOverflow()
		if overflow || uint64(len(ctx.childReturnData)) < end64 {
			return
		}
		ctx.mem.mstoreChunk(memOffset.Uint64(), length.Uint64(), ctx.childReturnData[offset64:end64])
		ctx.sta.pop(3)
	case op == vm.EXTCODEHASH:
		stack := scope.Stack.Data()
		addr := types.BytesToAddress(stack[len(stack)-1].Bytes())

		ctx.sta.pop(1)
		ctx.sta.push(newWordWithNilAncestor(ExtCodeHash, NewAddressLocation(addr), 32))
	case op == vm.BLOCKHASH:
		ctx.sta.pop(1)
		ctx.sta.push(newWordWithNilAncestor(BlockHash, emptyLoc, 32))
	case op == vm.COINBASE:
		ctx.sta.push(newWordWithNilAncestor(Coinbase, emptyLoc, 20))
	case op == vm.TIMESTAMP:
		ctx.sta.push(newWordWithNilAncestor(BlockTimestamp, emptyLoc, 32))
	case op == vm.NUMBER:
		ctx.sta.push(newWordWithNilAncestor(BlockNumber, emptyLoc, 32))
	case op == vm.DIFFICULTY:
		ctx.sta.push(newWordWithNilAncestor(Difficulty, emptyLoc, 32))
	case op == vm.RANDOM:
		ctx.sta.push(newWordWithNilAncestor(Random, emptyLoc, 32))
	case op == vm.GASLIMIT:
		ctx.sta.push(newWordWithNilAncestor(GasLimit, emptyLoc, 32))
	case op == vm.CHAINID:
		ctx.sta.push(newWordWithNilAncestor(ChainId, emptyLoc, 32))
	case op == vm.SELFBALANCE:
		ctx.sta.push(bal.readBalance(ctx.addr, ctx.callId))
	case op == vm.BASEFEE:
		ctx.sta.push(newWordWithNilAncestor(BaseFee, emptyLoc, 32))
	case op == vm.POP:
		ctx.sta.pop(1)
	case op == vm.MLOAD:
		stack := scope.Stack.Data()
		offset := stack[len(stack)-1].Uint64()

		ctx.sta.pop(1)
		ctx.sta.push(ctx.mem.mload(offset))
	case op == vm.MSTORE:
		stack := scope.Stack.Data()
		mStart := stack[len(stack)-1].Uint64()
		val := ctx.sta.s[len(ctx.sta.s)-2]

		ctx.mem.mstore(mStart, val)
		ctx.sta.pop(2)
	case op == vm.MSTORE8:
		stack := scope.Stack.Data()
		off := stack[len(stack)-1].Uint64()
		val := ctx.sta.s[len(ctx.sta.s)-2]

		ctx.mem.mstore8(off, val[31])
		ctx.sta.pop(2)
	case op == vm.SLOAD:
		stack := scope.Stack.Data()
		loc := stack[len(stack)-1]
		hash := types.Bytes32ToHash(loc.Bytes32())

		slotSource := mergeDataSources(ctx.sta.s[len(ctx.sta.s)-1][:]...)

		val := stg.sload(ctx.addr, hash, slotSource, dcfg.GetCurrentDcfgNodeId())

		ctx.sta.pop(1)
		ctx.sta.push(val)
	case op == vm.SSTORE:
		stack := scope.Stack.Data()
		loc := stack[len(stack)-1]
		hash := types.Bytes32ToHash(loc.Bytes32())

		val := ctx.sta.s[len(ctx.sta.s)-2]
		slotSource := mergeDataSources(ctx.sta.s[len(ctx.sta.s)-1][:]...)
		hid := stg.sstore(ctx.addr, hash, slotSource, val, dcfg.GetCurrentDcfgNodeId())
		ctx.storageWrites = append(
			ctx.storageWrites,
			&StorageLocation{
				addr: ctx.addr,
				slot: hash,
				hid:  hid,
			},
		)
		ctx.sta.pop(2)
	case op == vm.JUMP:
		ctx.sta.pop(1)
	case op == vm.JUMPI:
		// TODO the data doesn't flow directly through JUMP but it controls the code branches
		ctx.sta.pop(2)
	case op == vm.PC:
		ctx.sta.push(newWordWithNilAncestor(Pc, emptyLoc, 32))
	case op == vm.MSIZE:
		ctx.sta.push(newWordWithNilAncestor(MemorySize, emptyLoc, 32))
	case op == vm.GAS:
		ctx.sta.push(newWordWithNilAncestor(GasLeft, emptyLoc, 32))
	case op == vm.JUMPDEST:
	case op.IsPush():
		val := newCode(ctx.codeAddr, pc+1, uint64(op-vm.PUSH1)+1)
		var w word
		copy(w[32-len(val):], val)
		ctx.sta.push(w)
	case op == vm.PUSH0:
		ctx.sta.push(newNilWord())
	case uint64(op) >= uint64(vm.DUP1) && uint64(op) <= uint64(vm.DUP16):
		ctx.sta.dup(int(uint64(op) - uint64(vm.DUP1) + 1))
	case uint64(op) >= uint64(vm.SWAP1) && uint64(op) <= uint64(vm.SWAP16):
		ctx.sta.swap(int(uint64(op) - uint64(vm.SWAP1) + 1))
	case uint64(op) >= uint64(vm.LOG0) && uint64(op) <= uint64(vm.LOG4):
		// stack := scope.Stack.Data()
		// mStart := int(stack[len(stack)-1].Uint64())
		// mSize := int(stack[len(stack)-2].Uint64())
		// size := uint64(op) - uint64(vm.LOG0)
		// TODO consider if we need to capture LOG when connecting to the asset flow tracer
		ctx.sta.pop(int(uint64(op)-uint64(vm.LOG0)) + 2)
	case op == vm.CREATE:
		ctx.childCallValue = ctx.sta.s[len(ctx.sta.s)-1]
		ctx.sta.pop(3)
	case op == vm.CALL:
		stack := scope.Stack.Data()
		argsOffset := stack[len(stack)-4].Uint64()
		argsLength := stack[len(stack)-5].Uint64()
		retOffset := stack[len(stack)-6].Uint64()
		retLength := stack[len(stack)-7].Uint64()

		ctx.childCallValue = ctx.sta.s[len(ctx.sta.s)-3]
		ctx.childCallData = make([]*DataSource, argsLength)
		copy(ctx.childCallData, ctx.mem.mloadChunk(argsOffset, argsLength))

		ctx.retOffset = retOffset
		ctx.retLength = retLength

		ctx.sta.pop(7)

		// TODO take care of precompile contracts
		to := new(big.Int).SetBytes(stack[len(stack)-2].Bytes())
		if to.Cmp(big.NewInt(1)) >= 0 && to.Cmp(big.NewInt(9)) <= 0 {
			ctx.sta.push(newNilWord())
		}
	case op == vm.CALLCODE:
		stack := scope.Stack.Data()
		argsOffset := stack[len(stack)-4].Uint64()
		argsLength := stack[len(stack)-5].Uint64()
		retOffset := stack[len(stack)-6].Uint64()
		retLength := stack[len(stack)-7].Uint64()

		ctx.childCallValue = ctx.sta.s[len(ctx.sta.s)-3]
		ctx.childCallData = make([]*DataSource, argsLength)
		copy(ctx.childCallData, ctx.mem.mloadChunk(argsOffset, argsLength))

		ctx.retOffset = retOffset
		ctx.retLength = retLength

		ctx.sta.pop(7)

		// TODO take care of precompile contracts
		to := new(big.Int).SetBytes(stack[len(stack)-2].Bytes())
		if to.Cmp(big.NewInt(1)) >= 0 && to.Cmp(big.NewInt(9)) <= 0 {
			ctx.sta.push(newNilWord())
		}
	case op == vm.RETURN:
		stack := scope.Stack.Data()
		offset := stack[len(stack)-1].Uint64()
		size := stack[len(stack)-2].Uint64()

		ctx.returnData = newReturnData(ctx.callId, ctx.mem.mloadChunk(offset, size))

		ctx.sta.pop(2)
	case op == vm.DELEGATECALL:
		stack := scope.Stack.Data()
		argsOffset := stack[len(stack)-3].Uint64()
		argsLength := stack[len(stack)-4].Uint64()
		retOffset := stack[len(stack)-5].Uint64()
		retLength := stack[len(stack)-6].Uint64()

		ctx.childCallValue = ctx.callValue
		ctx.childCallData = make([]*DataSource, argsLength)
		copy(ctx.childCallData, ctx.mem.mloadChunk(argsOffset, argsLength))

		ctx.retOffset = retOffset
		ctx.retLength = retLength

		ctx.sta.pop(6)

		// TODO take care of precompile contracts
		to := new(big.Int).SetBytes(stack[len(stack)-2].Bytes())
		if to.Cmp(big.NewInt(1)) >= 0 && to.Cmp(big.NewInt(9)) <= 0 {
			ctx.sta.push(newNilWord())
		}
	case op == vm.CREATE2:
		ctx.childCallValue = ctx.sta.s[len(ctx.sta.s)-1]
		ctx.sta.pop(4)
	case op == vm.STATICCALL:
		stack := scope.Stack.Data()
		argsOffset := stack[len(stack)-3].Uint64()
		argsLength := stack[len(stack)-4].Uint64()
		retOffset := stack[len(stack)-5].Uint64()
		retLength := stack[len(stack)-6].Uint64()

		ctx.childCallValue = newNilWord()
		ctx.childCallData = make([]*DataSource, argsLength)
		copy(ctx.childCallData, ctx.mem.mloadChunk(argsOffset, argsLength))

		ctx.retOffset = retOffset
		ctx.retLength = retLength

		ctx.sta.pop(6)
		// TODO take care of precompile contracts
		to := new(big.Int).SetBytes(stack[len(stack)-2].Bytes())
		if to.Cmp(big.NewInt(1)) >= 0 && to.Cmp(big.NewInt(9)) <= 0 {
			ctx.sta.push(newNilWord())
		}
	case op == vm.REVERT:
		stack := scope.Stack.Data()
		offset := stack[len(stack)-1].Uint64()
		size := stack[len(stack)-2].Uint64()

		ctx.returnData = newReturnData(ctx.callId, ctx.mem.mloadChunk(offset, size))

		ctx.sta.pop(2)
	case op == vm.SELFDESTRUCT:
		ctx.childCallValue = bal.readBalance(ctx.addr, ctx.callId)
		ctx.sta.pop(1)
	case op == vm.STOP:
	default:
		panic(fmt.Errorf("this op code (%s) is missing", op.String()))
	}
}
