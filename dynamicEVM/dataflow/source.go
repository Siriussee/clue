package dataflow

import (
	"fmt"
	"sort"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
)

type SourceType uint8

const (
	Address SourceType = iota
	Balance
	Origin
	Caller
	CallValue
	CallData // location info
	CallDataSize
	CodeSize
	Code // location info
	GasPrice
	ExtCodeSize
	// ExtCode    // location info
	ReturnData // location info
	ReturnDataSize
	ExtCodeHash
	BlockHash
	Coinbase
	BlockTimestamp
	BlockNumber
	Difficulty
	Random
	GasLimit
	ChainId
	SelfBalance
	BaseFee
	Pc
	MemorySize
	GasLeft
	Storage // location info
	ExecutionResult
)

var sourceTypeToString = map[SourceType]string{
	Address:      "Address",
	Balance:      "Balance",
	Origin:       "Origin",
	Caller:       "Caller",
	CallValue:    "CallValue",
	CallData:     "CallData",
	CallDataSize: "CallDataSize",
	CodeSize:     "CodeSize",
	Code:         "Code",
	GasPrice:     "GasPrice",
	ExtCodeSize:  "ExtCodeSize",
	// ExtCode:         "ExtCode",
	ReturnData:      "ReturnData",
	ReturnDataSize:  "ReturnDataSize",
	ExtCodeHash:     "ExtCodeHash",
	BlockHash:       "BlockHash",
	Coinbase:        "Coinbase",
	BlockTimestamp:  "BlockTimestamp",
	BlockNumber:     "BlockNumber",
	Difficulty:      "Difficulty",
	Random:          "Random",
	GasLimit:        "GasLimit",
	ChainId:         "ChainId",
	SelfBalance:     "SelfBalance",
	BaseFee:         "BaseFee",
	Pc:              "Pc",
	MemorySize:      "MemorySize",
	GasLeft:         "GasLeft",
	Storage:         "Storage",
	ExecutionResult: "ExecutionResult",
}

func (typ SourceType) String() string {
	return sourceTypeToString[typ]
}

type Source struct {
	Typ      SourceType
	Loc      DataSourceLocation
	Ancestor *DataSource
}

type DataSource struct {
	Sources map[SourceType][]*Source
}

func (s *DataSource) String() string {
	var output string
	for typ := range s.Sources {
		output += fmt.Sprintf("%s\n", typ)
		switch typ {
		case CallData, ReturnData:
			data := make(map[string][]int)
			for _, s := range s.Sources[typ] {
				callIdStr := s.Loc.(*IODataLocation).callId.String()
				if _, callExists := data[callIdStr]; !callExists {
					data[callIdStr] = []int{}
				}
				data[callIdStr] = append(data[callIdStr], s.Loc.(*IODataLocation).pos)
			}
			for call, indices := range data {
				output += fmt.Sprintf("%s %s\n", call, indicesToRange(indices))
			}
		case Code:
			codes := make(map[types.Address][]int)
			for _, s := range s.Sources[typ] {
				loc := s.Loc.(*CodeLocation)
				if _, addrExists := codes[loc.addr]; !addrExists {
					codes[loc.addr] = []int{}
				}
				codes[loc.addr] = append(codes[loc.addr], loc.pos)
			}
			for addr, indices := range codes {
				output += fmt.Sprintf("%s %s\n", types.AddressToHex(addr), indicesToRange(indices))
			}
		case Storage:
			for _, s := range s.Sources[typ] {
				loc := s.Loc.(*StorageLocation)
				output += fmt.Sprintf("%s %s %d\n", types.AddressToHex(loc.addr), loc.slot.Hex(), loc.hid)
			}
		}
	}
	return output
}

func indicesToRange(indices []int) dataRanges {
	var (
		rs dataRanges = []*dataRange{}
		r  *dataRange
	)

	sorted := make([]int, len(indices))
	copy(sorted, indices)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	for _, i := range sorted {
		if r == nil {
			r = &dataRange{
				start: i,
				end:   i + 1,
			}
			rs = append(rs, r)
		} else if i == r.end {
			r.end += 1
		} else {
			r = &dataRange{
				start: i,
				end:   i + 1,
			}
			rs = append(rs, r)
		}
	}

	return rs
}

type dataRange struct {
	start int
	end   int
}

type dataRanges []*dataRange

func (dr dataRanges) String() string {
	var output string

	output += "["
	for _, r := range dr {
		output += fmt.Sprintf("%d: %d, ", r.start, r.end)
	}
	output += "]"

	return output
}

type DataSourceLocation interface {
	String() string
}

type IODataLocation struct {
	callId types.CallId
	pos    int
}

func NewIODataLocation(callId types.CallId, pos int) *IODataLocation {
	return &IODataLocation{
		callId: callId,
		pos:    pos,
	}
}

func (loc *IODataLocation) CallId() types.CallId {
	return loc.callId
}

func (loc *IODataLocation) String() string {
	return fmt.Sprintf("%s:%d", loc.callId, loc.pos)
}

type CodeLocation struct {
	addr types.Address
	pos  int
}

func NewCodeLocation(addr types.Address, pos int) *CodeLocation {
	return &CodeLocation{
		addr: addr,
		pos:  pos,
	}
}

func (loc *CodeLocation) Address() types.Address {
	return loc.addr
}

func (loc *CodeLocation) String() string {
	return fmt.Sprintf("%s:%d", types.AddressToHex(loc.addr), loc.pos)
}

type StorageLocation struct {
	addr types.Address
	slot types.Hash
	hid  int
}

func (loc *StorageLocation) Address() types.Address {
	return loc.addr
}

func (loc *StorageLocation) Slot() types.Hash {
	return loc.slot
}

func (loc *StorageLocation) HistoryId() int {
	return loc.hid
}

func NewStorageLocation(addr types.Address, slot types.Hash, hid int) *StorageLocation {
	return &StorageLocation{
		addr: addr,
		slot: slot,
		hid:  hid,
	}
}

func (loc *StorageLocation) String() string {
	return fmt.Sprintf("%s:%s:%d", types.AddressToHex(loc.addr), loc.slot.Hex(), loc.hid)
}

type addressLocation struct {
	addr types.Address
}

func NewAddressLocation(addr types.Address) *addressLocation {
	return &addressLocation{
		addr: addr,
	}
}

func (loc *addressLocation) String() string {
	return types.AddressToHex(loc.addr)
}

type BalanceLocation struct {
	addr types.Address
	hid  int
}

func (loc *BalanceLocation) Address() types.Address {
	return loc.addr
}

func (loc *BalanceLocation) HistoryId() int {
	return loc.hid
}

func NewBalanceLocation(addr types.Address, hid int) *BalanceLocation {
	return &BalanceLocation{
		addr: addr,
		hid:  hid,
	}
}

func (loc *BalanceLocation) String() string {
	return fmt.Sprintf("%s:%d", types.AddressToHex(loc.addr), loc.hid)
}

func MergeSources(sources ...*Source) []*Source {
	merged := []*Source{}
	switch sources[0].Loc.(type) {
	case *IODataLocation, *StorageLocation, *BalanceLocation:
		tracker := make(map[*Source]struct{})
		for _, s := range sources {
			if _, exists := tracker[s]; !exists {
				tracker[s] = struct{}{}
				merged = append(merged, s)
			}
		}

	case *CodeLocation:
		tracker := make(map[types.Address]map[int]struct{})
		for _, s := range sources {
			loc := s.Loc.(*CodeLocation)
			if _, exists := tracker[loc.addr]; !exists {
				tracker[loc.addr] = make(map[int]struct{})
			}
			if _, exists := tracker[loc.addr][loc.pos]; !exists {
				tracker[loc.addr][loc.pos] = struct{}{}
				merged = append(merged, s)
			}
		}

	case *addressLocation:
		tracker := make(map[types.Address]struct{})
		for _, s := range sources {
			loc := s.Loc.(*addressLocation)
			if _, exists := tracker[loc.addr]; !exists {
				tracker[loc.addr] = struct{}{}
				merged = append(merged, s)
			}
		}

	case types.CallId:
		tracker := make(map[string]struct{})
		for _, s := range sources {
			loc := s.Loc.(types.CallId).String()
			if _, exists := tracker[loc]; !exists {
				tracker[loc] = struct{}{}
				merged = append(merged, s)
			}
		}

	case nil:
		merged = append(merged, sources[0])

	default:
		panic("MergeSources: unknown sources")
	}

	return merged
}

var emptyLoc DataSourceLocation = nil

func newSource(typ SourceType, loc DataSourceLocation, ancestor *DataSource) *Source {
	return &Source{
		Typ:      typ,
		Loc:      loc,
		Ancestor: ancestor,
	}
}

func newDataSource(typ SourceType, loc DataSourceLocation, ancestor *DataSource) *DataSource {
	return &DataSource{
		Sources: map[SourceType][]*Source{
			typ: {newSource(typ, loc, ancestor)},
		},
	}
}

func rightPadNil(slice []*DataSource, l uint64) []*DataSource {
	if int(l) <= len(slice) {
		return slice
	}

	padded := make([]*DataSource, l)
	copy(padded, slice)

	return padded
}

func getData(data []*DataSource, start uint64, size uint64) []*DataSource {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return rightPadNil(data[start:end], size)
}

type word [32]*DataSource

func newNilWord() word {
	return [32]*DataSource{}
}

func newWordWithNilAncestor(typ SourceType, loc DataSourceLocation, size uint64) word {
	if size > 32 {
		panic("stack newWord: invalid size")
	}
	var w word

	s := newDataSource(typ, loc, nil)
	for i := uint64(0); i < size; i++ {
		w[32-size+i] = s
	}
	return w
}

// call data
func newCallData(callId types.CallId, ancestors []*DataSource) []*DataSource {
	calldata := make([]*DataSource, len(ancestors))

	for i := 0; i < len(ancestors); i++ {
		calldata[i] = &DataSource{
			Sources: map[SourceType][]*Source{
				CallData: {newSource(CallData, NewIODataLocation(callId, i), ancestors[i])},
			},
		}
	}

	return calldata
}

// return data
func newReturnData(callId types.CallId, ancestors []*DataSource) []*DataSource {
	calldata := make([]*DataSource, len(ancestors))

	for i := 0; i < len(ancestors); i++ {
		calldata[i] = &DataSource{
			Sources: map[SourceType][]*Source{
				ReturnData: {newSource(ReturnData, NewIODataLocation(callId, i), ancestors[i])},
			},
		}
	}

	return calldata
}

// execution result
func newExecResult(callId types.CallId) word {
	var w word
	w[31] = &DataSource{
		Sources: map[SourceType][]*Source{
			ExecutionResult: {newSource(ExecutionResult, callId, nil)},
		},
	}
	return w
}

// code
func newCode(addr types.Address, start uint64, size uint64) []*DataSource {
	code := make([]*DataSource, size)

	for i := uint64(0); i < size; i++ {
		code[i] = newDataSource(Code, NewCodeLocation(addr, int(start+i)), nil)
	}

	return code
}

// storage
func newStorageWord(addr types.Address, slot types.Hash, hid int, ancestor word) word {
	w := newNilWord()

	mergedAncestor := mergeDataSources(ancestor[:]...)
	s := newDataSource(Storage, NewStorageLocation(addr, slot, hid), mergedAncestor)

	for i := 0; i < 32; i++ {
		w[i] = s
	}

	return w
}

// balance
func newBalanceWord(addr types.Address, hid int, ancestor word) word {
	w := newNilWord()

	mergedAncestor := mergeDataSources(ancestor[:]...)
	s := newDataSource(Balance, NewBalanceLocation(addr, hid), mergedAncestor)

	for i := 0; i < 32; i++ {
		w[i] = s
	}

	return w
}

// call value
func newCallValue(callId types.CallId, ancestor word) word {
	w := newNilWord()

	for i := 0; i < 32; i++ {
		s := newDataSource(CallValue, callId, ancestor[i])
		w[i] = s
	}

	return w
}

func mergeDataSources(dataSources ...*DataSource) *DataSource {
	filtered := make(map[SourceType]map[*Source]struct{})
	for _, dataSource := range dataSources {
		if dataSource == nil {
			continue
		}
		for dataType := range dataSource.Sources {
			if _, typeExists := filtered[dataType]; !typeExists {
				filtered[dataType] = make(map[*Source]struct{})
			}
			for _, source := range dataSource.Sources[dataType] {
				if source != nil {
					filtered[dataType][source] = struct{}{}
				}
			}
		}
	}

	sources := make(map[SourceType][]*Source)

	for dataType := range filtered {
		toMerge := []*Source{}
		for s := range filtered[dataType] {
			toMerge = append(toMerge, s)
		}
		sources[dataType] = MergeSources(toMerge...)
	}

	return &DataSource{
		Sources: sources,
	}

}

func byteMerge(x *DataSource, y *DataSource) *DataSource {
	if x == nil && y == nil {
		return nil
	} else if x != nil && y == nil {
		return x
	} else if x == nil && y != nil {
		return y
	} else {
		return mergeDataSources(x, y)
	}
}

func wordMerge(x word, y word) word {
	var w word
	for i := 0; i < 32; i++ {
		if x[i] != nil && y[i] != nil {
			w[i] = byteMerge(x[i], y[i])
		} else if x[i] != nil && y[i] == nil {
			w[i] = x[i]
		} else if x[i] == nil && y[i] != nil {
			w[i] = y[i]
		}
	}
	return w
}
