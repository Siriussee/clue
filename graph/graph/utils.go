package graph

import (
	"github.com/anon/execution-property-graph/dynamicEVM/dataflow"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
)

func dataflowSourceIdentifier(source *dataflow.Source) string {
	return source.Typ.String() + "#" + dataflowSourceLocationToString(source.Loc)
}

func dataflowSourceLocationToString(location dataflow.DataSourceLocation) string {
	if location == nil {
		return "emptyLoc"
	}
	switch location := location.(type) {
	case *dataflow.IODataLocation:
		return location.CallId().String() // Merge pos in IOData
	case *dataflow.CodeLocation:
		return types.AddressToHex(location.Address()) // Merge pos in Code
	}
	return location.String()
}
