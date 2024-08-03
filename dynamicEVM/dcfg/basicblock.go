package dcfg

import (
	"bytes"
	"errors"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/dynamicEVM/vm"
)

const INVALIDPC = types.MaxUint64

type BasicBlockType int64

const (
	Runtime BasicBlockType = iota
	Creation
)

var BasicBlockEnd = map[vm.OpCode]struct{}{
	vm.STOP:         {},
	vm.SELFDESTRUCT: {},
	vm.RETURN:       {},
	vm.REVERT:       {},
	vm.JUMP:         {},
	vm.JUMPI:        {},
}

var EnterOp = map[vm.OpCode]struct{}{
	vm.CALL:         {},
	vm.CALLCODE:     {},
	vm.DELEGATECALL: {},
	vm.STATICCALL:   {},
	vm.CREATE:       {},
	vm.CREATE2:      {},
	vm.SELFDESTRUCT: {},
}

type BasicBlock struct {
	blockType BasicBlockType
	codeAddr  types.Address
	pc        uint64
	code      []byte
}

func NewBasicBlock(blockType BasicBlockType, codeAddr types.Address) *BasicBlock {
	return &BasicBlock{
		blockType: blockType,
		codeAddr:  codeAddr,
		pc:        INVALIDPC,
		code:      []byte{},
	}
}

func (b *BasicBlock) Address() types.Address {
	return b.codeAddr
}

func (b *BasicBlock) Pc() uint64 {
	return b.pc
}

func (b *BasicBlock) Code() []byte {
	return b.code
}

func (b *BasicBlock) Len() uint64 {
	return uint64(len(b.code))
}

func (b *BasicBlock) IsJump() bool {
	return vm.OpCode(b.code[len(b.code)-1]) == vm.JUMP
}

func (b *BasicBlock) IsJumpi() bool {
	return vm.OpCode(b.code[len(b.code)-1]) == vm.JUMPI
}

func (b *BasicBlock) Snippet(pc uint64, length uint64) []byte {
	if pc < b.pc || pc+length > b.pc+uint64(len(b.code)) {
		return nil
	}
	return b.code[pc-b.pc : pc-b.pc+length]
}

func (b *BasicBlock) addInstructions(pc uint64, instructions []byte) error {
	if b.pc == INVALIDPC {
		b.pc = pc
	}

	if pc == b.pc+uint64(len(b.code)) {
		b.code = append(b.code, instructions...)
	} else if pc >= b.pc && pc+uint64(len(instructions)) <= b.pc+uint64(len(b.code)) {
		if !bytes.Equal(instructions, b.code[pc-b.pc:pc-b.pc+uint64(len(instructions))]) {
			return errors.New("inconsistent basic block instructions")
		}
	} else {
		return errors.New("inconsistent basic block length")
	}

	return nil
}
