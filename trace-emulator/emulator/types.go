package emulator

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/holiman/uint256"
)

// ExecutionResult groups all structured logs emitted by the EVM
// while replaying a transaction in debug mode as well as transaction
// execution status, the amount of gas used and the return value
type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

type EthError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode
type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   *string            `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

type CallFrame struct {
	Type    string      `json:"type"`
	From    string      `json:"from"`
	To      string      `json:"to,omitempty"`
	Value   string      `json:"value,omitempty"`
	Gas     string      `json:"gas"`
	GasUsed string      `json:"gasUsed"`
	Input   string      `json:"input"`
	Output  string      `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
	Calls   []CallFrame `json:"calls,omitempty"`
}

// BSON
type ExecutionResultBSON struct {
	TxHash      [32]byte
	Gas         uint64
	Failed      bool
	ReturnValue []byte
	StructLogs  []StructLogResBSON `bson:"omitempty"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode
type StructLogResBSON struct {
	TxHash  [32]byte
	Index   int
	Pc      uint64
	Op      string
	Gas     uint64
	GasCost uint64
	Depth   int
	Error   bool
	Stack   [][32]byte
	Memory  []byte
	Storage map[string]string
}

type CallFrameBSON struct {
	Type    string
	From    []byte
	To      []byte
	Value   []byte
	Gas     uint64
	GasUsed uint64
	Input   []byte
	Output  []byte
	Error   string          `bson:"error,omitempty"`
	Calls   []CallFrameBSON `bson:"calls,omitempty"`
}

type TxCallFrame struct {
	TxHash    [32]byte
	CallFrame CallFrameBSON
}

func (res ExecutionResult) ToBSON(txHash [32]byte) (ExecutionResultBSON, []StructLogResBSON, error) {
	var structLogs []StructLogResBSON
	for i, structLog := range res.StructLogs {
		structLogBSON := structLog.ToBSON(txHash, i)
		structLogs = append(structLogs, structLogBSON)
	}

	var (
		returnValue []byte
		decodeErr   error
	)
	if res.ReturnValue == "" {
		returnValue = []byte{}
	} else {
		returnValue, decodeErr = hex.DecodeString(res.ReturnValue[2:])
		if decodeErr != nil {
			err := fmt.Errorf("decode output %s error: %w", res.ReturnValue, decodeErr)
			return ExecutionResultBSON{}, []StructLogResBSON{}, err
		}
	}

	return ExecutionResultBSON{
		TxHash:      txHash,
		Gas:         res.Gas,
		Failed:      res.Failed,
		ReturnValue: returnValue,
	}, structLogs, nil
}

func (res StructLogRes) ToBSON(txHash [32]byte, index int) StructLogResBSON {
	var (
		resError   bool
		resStack   [][32]byte
		resStorage map[string]string
	)

	if res.Error != nil {
		resError = true
	} else {
		resError = false
	}

	stack := newStack(res.Stack)
	for _, s := range stack.data {
		resStack = append(resStack, s.Bytes32())
	}

	if res.Storage == nil {
		resStorage = nil
	} else {
		resStorage = *res.Storage
	}

	return StructLogResBSON{
		TxHash:  txHash,
		Index:   index,
		Pc:      res.Pc,
		Op:      res.Op,
		Gas:     res.Gas,
		GasCost: res.GasCost,
		Depth:   res.Depth,
		Error:   resError,
		Stack:   resStack,
		Memory:  newMemory(res.Memory).data,
		Storage: resStorage,
	}
}

func (frame CallFrame) ToBSON() (CallFrameBSON, error) {
	var calls []CallFrameBSON
	for _, subCall := range frame.Calls {
		subCallBson, err := subCall.ToBSON()
		if err != nil {
			return CallFrameBSON{}, err
		}
		calls = append(calls, subCallBson)
	}

	var (
		value   *big.Int
		gas     uint64
		gasUsed uint64
		input   []byte
		output  []byte
	)

	if frame.Value != "" {
		var decoded bool
		value, decoded = new(big.Int).SetString(frame.Value, 0)
		if !decoded {
			err := fmt.Errorf("decode value %s error", frame.Value)
			return CallFrameBSON{}, err
		}
	} else {
		value = big.NewInt(0)
	}

	var decodeErr error
	if frame.Gas == "" {
		gas = 0
	} else {
		gas, decodeErr = strconv.ParseUint(frame.Gas[2:], 16, 64)
		if decodeErr != nil {
			err := fmt.Errorf("decode gas %s error: %w", frame.Gas, decodeErr)
			return CallFrameBSON{}, err
		}
	}
	if frame.GasUsed == "" {
		gasUsed = 0
	} else {
		gasUsed, decodeErr = strconv.ParseUint(frame.GasUsed[2:], 16, 64)
		if decodeErr != nil {
			err := fmt.Errorf("decode gasUsed %s error: %w", frame.GasUsed, decodeErr)
			return CallFrameBSON{}, err
		}
	}
	if frame.Input == "" {
		input = []byte{}
	} else {
		input, decodeErr = hex.DecodeString(frame.Input[2:])
		if decodeErr != nil {
			err := fmt.Errorf("decode input %s error: %w", frame.Input, decodeErr)
			return CallFrameBSON{}, err
		}
	}
	if frame.Output == "" {
		output = []byte{}
	} else {
		output, decodeErr = hex.DecodeString(frame.Output[2:])
		if decodeErr != nil {
			err := fmt.Errorf("decode output %s error: %w", frame.Output, decodeErr)
			return CallFrameBSON{}, err
		}
	}

	return CallFrameBSON{
		Type:    frame.Type,
		From:    types.HexToAddress(frame.From).Bytes(),
		To:      types.HexToAddress(frame.To).Bytes(),
		Value:   value.Bytes(),
		Gas:     gas,
		GasUsed: gasUsed,
		Input:   input,
		Output:  output,
		Error:   frame.Error,
		Calls:   calls,
	}, nil
}

func (frame *CallFrame) decode() (from types.Address, to types.Address, value *big.Int, gas uint64, gasUsed uint64, input []byte, output []byte, err error) {
	from = types.HexToAddress(frame.From)
	to = types.HexToAddress(frame.To)
	value = new(big.Int)
	if frame.Value != "" {
		var decoded bool
		value, decoded = new(big.Int).SetString(frame.Value, 0)
		if !decoded {
			err = fmt.Errorf("decode value %s error", frame.Value)
			return
		}
	}
	var decodeErr error
	if frame.Gas == "" {
		gas = 0
	} else {
		gas, decodeErr = strconv.ParseUint(frame.Gas[2:], 16, 64)
		if decodeErr != nil {
			err = fmt.Errorf("decode gas %s error: %w", frame.Gas, decodeErr)
			return
		}
	}
	if frame.GasUsed == "" {
		gasUsed = 0
	} else {
		gasUsed, decodeErr = strconv.ParseUint(frame.GasUsed[2:], 16, 64)
		if err != nil {
			err = fmt.Errorf("decode gasUsed %s error: %w", frame.GasUsed, decodeErr)
			return
		}
	}
	if frame.Input == "" {
		input = []byte{}
	} else {
		input, decodeErr = hex.DecodeString(frame.Input[2:])
		if decodeErr != nil {
			err = fmt.Errorf("decode input %s error: %w", frame.Input, decodeErr)
			return
		}
	}
	if frame.Output == "" {
		output = []byte{}
	} else {
		output, decodeErr = hex.DecodeString(frame.Output[2:])
		if decodeErr != nil {
			err = fmt.Errorf("decode output %s error: %w", frame.Output, decodeErr)
			return
		}
	}
	return
}

func (frame *CallFrameBSON) decode() (from types.Address, to types.Address, value *big.Int, gas uint64, gasUsed uint64, input []byte, output []byte) {
	from = types.BytesToAddress(frame.From)
	to = types.BytesToAddress(frame.To)
	value = new(big.Int).SetBytes(frame.Value)
	gas = frame.Gas
	gasUsed = frame.GasUsed
	input = frame.Input
	output = frame.Output

	return
}

type callCursor struct {
	frame *CallFrame
	addr  types.Address
	index int
	rData []byte
}

type callCursorBSON struct {
	frame *CallFrameBSON
	addr  types.Address
	index int
	rData []byte
}

type memory struct {
	data []byte
}

func newMemory(m *[]string) *memory {
	data := []byte{}
	for _, str := range *m {
		b, err := hex.DecodeString(str)
		if err != nil {
			panic(fmt.Errorf("unable to decode %s: %w", str, err))
		}
		if len(b) != 32 {
			panic(fmt.Sprintf("decode memory %s length error: %d", str, len(b)))
		}
		data = append(data, b...)
	}

	return &memory{data: data}
}

func cpMemory(m []byte) *memory {
	data := make([]byte, len(m))
	copy(data, m)

	return &memory{data: data}
}

func (m *memory) Data() []byte {
	return m.data
}

type stack struct {
	data []uint256.Int
}

func newStack(s *[]string) *stack {
	data := []uint256.Int{}

	for _, str := range *s {
		// Some traces will have odd hex like "0x4"
		h := str
		if strings.HasPrefix(str, "0x") {
			h = str[2:]
		}
		if len(h)%2 != 0 {
			h = "0" + h
		}
		b, err := hex.DecodeString(h)
		if err != nil {
			panic(fmt.Errorf("unable to decode %s: %w", str, err))
		}
		data = append(data, *(uint256.NewInt(0).SetBytes(b)))
	}

	return &stack{data: data}
}

func cpStack(s [][32]byte) *stack {
	data := []uint256.Int{}
	for _, word := range s {
		data = append(data, *(uint256.NewInt(0).SetBytes(word[:])))
	}

	return &stack{data: data}
}

func (s *stack) Data() []uint256.Int {
	return s.data
}

func (s *stack) Manipulate(int, uint256.Int) {
	panic("manipulate not allowed for trace emulator")
}

func newContract(addr types.Address) *types.Contract {
	return &types.Contract{
		Address: addr,
	}
}
