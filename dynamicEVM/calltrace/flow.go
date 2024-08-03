package calltrace

import (
	"bytes"
	"math/big"

	"github.com/anon/execution-property-graph/dynamicEVM/dataflow"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
)

type Flow struct {
	From   types.Address
	To     types.Address
	Asset  types.Address
	Amount *big.Int
	Index  int
	DcfgId types.DcfgId

	AmountTracker *dataflow.DataSource
}

func (f *Flow) IsEther() bool {
	return bytes.Equal(f.Asset.Bytes(), types.NativeTokenAddress.Bytes())
}

func NewEtherFlow(from, to types.Address, amount *big.Int, index int, dcfgId types.DcfgId, amountTracker *dataflow.DataSource) *Flow {
	return &Flow{
		From:          from,
		To:            to,
		Asset:         types.NativeTokenAddress,
		Amount:        new(big.Int).Set(amount),
		Index:         index,
		DcfgId:        dcfgId,
		AmountTracker: amountTracker,
	}
}

func NewERC20Flow(from, to, asset types.Address, amount *big.Int, index int, dcfgId types.DcfgId, amountTracker *dataflow.DataSource) *Flow {
	return &Flow{
		From:          from,
		To:            to,
		Asset:         asset,
		Amount:        new(big.Int).Set(amount),
		Index:         index,
		DcfgId:        dcfgId,
		AmountTracker: amountTracker,
	}
}

type Flows []*Flow
