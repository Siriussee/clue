package dataflow

import "github.com/anon/execution-property-graph/dynamicEVM/types"

type BalanceValue struct {
	addr   types.Address
	hid    int
	source word
	reads  []types.CallId
	write  types.CallId
}

func (b *BalanceValue) GetSources() []*DataSource {
	return b.source[:]
}

func (b *BalanceValue) GetReads() []types.CallId {
	return b.reads
}

func (b *BalanceValue) GetWrite() types.CallId {
	return b.write
}

type balance struct {
	b map[types.Address][]*BalanceValue
}

func newBalance() *balance {
	return &balance{
		b: make(map[types.Address][]*BalanceValue),
	}
}

func (b *balance) readBalance(addr types.Address, callId types.CallId) word {
	if _, balanceExists := b.b[addr]; balanceExists {
		currentValue := b.b[addr][len(b.b[addr])-1]
		currentValue.reads = append(currentValue.reads, callId)

		return currentValue.source
	} else {
		w := newBalanceWord(addr, 0, newNilWord())
		b.b[addr] = []*BalanceValue{
			{
				addr:   addr,
				hid:    0,
				source: w,
				reads:  []types.CallId{callId},
			},
		}

		return w
	}
}

func (b *balance) writeBalance(addr types.Address, value word, callId types.CallId) int {
	if _, balanceExists := b.b[addr]; balanceExists {
		hid := len(b.b[addr])
		bNew := &BalanceValue{
			addr:   addr,
			hid:    hid,
			source: newBalanceWord(addr, hid, value),
			reads:  []types.CallId{},
			write:  callId,
		}
		b.b[addr] = append(b.b[addr], bNew)
		return hid
	} else {
		bNew := &BalanceValue{
			addr:   addr,
			hid:    0,
			source: newBalanceWord(addr, 0, value),
			reads:  []types.CallId{},
			write:  callId,
		}
		b.b[addr] = []*BalanceValue{bNew}
		return 0
	}
}

func (b *balance) revert(addr types.Address, hid int) {
	if _, addrExists := b.b[addr]; addrExists {
		if hid == 0 {
			delete(b.b, addr)
		} else if hid < len(b.b[addr]) {
			b.b[addr] = b.b[addr][:hid]
		}
	}
}
