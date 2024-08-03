package dataflow

import (
	"github.com/anon/execution-property-graph/dynamicEVM/types"
)

type StorageValue struct {
	addr       types.Address
	slot       types.Hash
	hid        int
	source     word
	slotSource *DataSource
	reads      []types.DcfgId
	write      types.DcfgId
}

func (sv *StorageValue) GetSources() []*DataSource {
	return sv.source[:]
}

func (sv *StorageValue) GetReads() []types.DcfgId {
	return sv.reads
}

func (sv *StorageValue) GetWrite() types.DcfgId {
	return sv.write
}

// s stores storage for every smart contract involved in a transaction
type storage struct {
	s map[types.Address]map[types.Hash][]*StorageValue
}

func newStorage() *storage {
	return &storage{
		s: make(map[types.Address]map[types.Hash][]*StorageValue),
	}
}

func (s *storage) sload(addr types.Address, slot types.Hash, slotSource *DataSource, dcfgId types.DcfgId) word {
	if _, addrExists := s.s[addr]; !addrExists {
		s.s[addr] = make(map[types.Hash][]*StorageValue)
	}
	if _, slotExists := s.s[addr][slot]; slotExists {
		currentValue := s.s[addr][slot][len(s.s[addr][slot])-1]
		currentValue.reads = append(currentValue.reads, dcfgId)

		return currentValue.source
	} else {
		w := newStorageWord(addr, slot, 0, newNilWord())
		s.s[addr][slot] = []*StorageValue{
			{
				addr:       addr,
				slot:       slot,
				hid:        0,
				source:     w,
				slotSource: slotSource,
				reads:      []types.DcfgId{dcfgId},
			},
		}

		return w
	}
}

func (s *storage) sstore(addr types.Address, slot types.Hash, slotSource *DataSource, value word, dcfgId types.DcfgId) int {
	if _, addrExists := s.s[addr]; !addrExists {
		s.s[addr] = make(map[types.Hash][]*StorageValue)
	}
	if _, slotExists := s.s[addr][slot]; slotExists {
		hid := len(s.s[addr][slot])
		sNew := &StorageValue{
			addr:       addr,
			slot:       slot,
			hid:        hid,
			source:     newStorageWord(addr, slot, hid, value),
			slotSource: slotSource,
			reads:      []types.DcfgId{},
			write:      dcfgId,
		}
		s.s[addr][slot] = append(s.s[addr][slot], sNew)
		return hid
	} else {
		sNew := &StorageValue{
			addr:       addr,
			slot:       slot,
			hid:        0,
			source:     newStorageWord(addr, slot, 0, value),
			slotSource: slotSource,
			reads:      []types.DcfgId{},
			write:      dcfgId,
		}
		s.s[addr][slot] = []*StorageValue{sNew}
		return 0
	}
}

func (s *storage) revert(addr types.Address, slot types.Hash, hid int) {
	if _, addrExists := s.s[addr]; addrExists {
		if _, slotExists := s.s[addr][slot]; slotExists {
			if hid == 0 {
				delete(s.s[addr], slot)
			} else if hid < len(s.s[addr][slot]) {
				s.s[addr][slot] = s.s[addr][slot][:hid]
			}
		}
	}
}
