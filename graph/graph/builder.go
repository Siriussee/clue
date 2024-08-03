package graph

import (
	"github.com/anon/execution-property-graph/dynamicEVM/calltrace"
	"github.com/anon/execution-property-graph/dynamicEVM/dataflow"
	"github.com/anon/execution-property-graph/dynamicEVM/dcfg"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/graph/nodes"
	mapset "github.com/deckarep/golang-set/v2"
)

type GraphBuilder struct {
	config *GraphConfig

	session *GraphFactorySession
}

func NewGraphBuilder(config *GraphConfig) (*GraphBuilder, error) {
	return &GraphBuilder{
		config: config,
	}, nil
}

func (gb *GraphBuilder) BuildGraph(traceResult *TraceResult) (*Graph, error) {
	g, err := NewEmptyGraph(gb.config)
	if err != nil {
		return nil, err
	}
	gb.session = NewGraphFactorySession(g, 20)
	err = gb.buildCallTraceGraph(g, traceResult)
	if err != nil {
		return nil, err
	}
	err = gb.buildDcfgGraph(g, traceResult)
	if err != nil {
		return nil, err
	}
	err = gb.buildDataflowGraph(g, traceResult)
	if err != nil {
		return nil, err
	}
	if err := gb.session.Submit(); err != nil {
		return nil, err
	}
	gb.session = nil
	return g, nil
}

func (gb *GraphBuilder) buildCallTraceGraph(g *Graph, traceResult *TraceResult) error {
	callTraces := traceResult.CallTraces

	accountSet := make(map[string]string)
	for _, callTrace := range callTraces {
		accountSet[callTrace.ID.String()] = types.AddressToHex(callTrace.To)
	}

	addrs := make([]types.Address, 0)
	callIds := make([]types.CallId, 0)
	for callId, addr := range accountSet {
		addrs = append(addrs, types.HexToAddress(addr))
		cid, _ := types.StringToCallId(callId)
		callIds = append(callIds, cid)
	}
	accounts, err := nodes.CreateContractCallNodes(g.g, addrs, callIds)
	if err != nil {
		return err
	}
	g.logger.Debug("CreateContractCallNodes count: ", len(accounts))

	for _, account := range accounts {
		g.nodeMap.RegisterContractCallNode(account)
	}

	// Create call graph
	for _, callTrace := range callTraces {
		from := g.nodeMap.GetContractCallNode(callTrace.ID.Parent())
		to := g.nodeMap.GetContractCallNode(callTrace.ID)
		if from != to {
			_, err := gb.session.CreateCallEdge(from, to, callTrace)
			if err != nil {
				return err
			}
		}

		if len(callTrace.Flows) > 0 {
			callIds = make([]types.CallId, len(callTrace.Flows))
			indices := make([]uint64, len(callTrace.Flows))
			flows := make([]*calltrace.Flow, len(callTrace.Flows))

			for i, flow := range callTrace.Flows {
				callIds[i] = flow.DcfgId.CallId()
				indices[i] = uint64(flow.Index)
				flows[i] = flow
			}
			assetFlowNodes, err := nodes.CreateAssetFlowNodes(g.g, callIds, indices, flows)
			if err != nil {
				return err
			}
			g.nodeMap.RegisterAssetFlowNodes(assetFlowNodes)

			for _, flowNode := range assetFlowNodes {
				// Create asset flow edge
				_, err := gb.session.CreateAssetFlowEdge(to, flowNode)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (gb *GraphBuilder) buildDcfgGraph(g *Graph, traceResult *TraceResult) error {
	calls := make([]*dcfg.ContractNode, 0)
	dcfgRoot := traceResult.GetDcfg().GetRoot()
	calls = append(calls, dcfgRoot)
	traversalIdx := 0
	for traversalIdx < len(calls) {
		curr := calls[traversalIdx]
		traversalIdx++
		calls = append(calls, curr.GetCalls()...)
	}

	for _, call := range calls {
		g.logger.Debug("Processing DCFG for call: ", call.GetCallId().String())
		callNode := g.nodeMap.GetContractCallNode(call.GetCallId())
		if callNode == nil {
			g.logger.Warn("Cannot find call node for call: ", call.GetCallId().String())
			continue
		}
		// Get all DCFG blocks
		dcfgBlocks := make([]*dcfg.DcfgNode, 0)
		curr := call.GetEntry()
		for curr != nil {
			dcfgBlocks = append(dcfgBlocks, curr)
			curr = curr.GetChild()
		}
		dcfgIds := make([]types.DcfgId, 0)
		pcs := make([]uint64, 0)
		for _, node := range dcfgBlocks {
			dcfgId := types.NewDcfgId(call.GetCallId(), node.GetId(), node.GetCallCount())
			dcfgIds = append(dcfgIds, dcfgId)
			pcs = append(pcs, uint64(node.GetPc()))
		}
		// Add a dummy dcfg node for ETH transfers
		if len(dcfgIds) == 0 {
			dcfgIds = append(dcfgIds, types.NewDcfgId(call.GetCallId(), 0, 0))
			pcs = append(pcs, 0)
		}
		// Create DCFG nodes in graph
		dcfgNodes, err := nodes.CreateDcfgNodes(g.g, dcfgIds, pcs)
		if err != nil {
			return err
		}
		// Map DCFG node ids to graph nodes
		for _, node := range dcfgNodes {
			g.nodeMap.RegisterDcfgNode(node)
		}
		// Create start edge
		_, err = gb.session.CreateDcfgStartEdge(callNode, dcfgNodes[0])
		if err != nil {
			return err
		}

		targetCallId := types.NewCallId(make([]int, len(call.GetCallId())+1))
		copy(targetCallId, call.GetCallId())

		// Create all remaining edges
		for i := 0; i < len(dcfgNodes); i++ {
			if i < len(dcfgNodes)-1 {
				_, err = gb.session.CreateDcfgEdge(dcfgNodes[i], dcfgNodes[i+1], dcfgBlocks[i].GetOpcode(), dcfgBlocks[i].GetCondition())
				if err != nil {
					return err
				}
			}
			// create dcfg call based on call count
			if i > 0 {
				for callCount := dcfgNodes[i-1].DcfgId.CallCount(); callCount < dcfgNodes[i].DcfgId.CallCount(); callCount++ {
					targetCallId[len(targetCallId)-1] = callCount
					targetCallNode := g.nodeMap.GetContractCallNode(targetCallId)
					if targetCallNode == nil {
						g.logger.Warn("Cannot find call node for call: ", targetCallId.String())
						continue
					}
					_, err = gb.session.CreateDcfgCallEdge(dcfgNodes[i-1], targetCallNode)
					if err != nil {
						return err
					}
				}
			}
		}

		// Create Ret edge and check if the last node has some calls
		lastDcfgNode := dcfgNodes[len(dcfgNodes)-1]
		callCount := lastDcfgNode.DcfgId.CallCount()
		for true {
			targetCallId[len(targetCallId)-1] = callCount
			targetCallNode := g.nodeMap.GetContractCallNode(targetCallId)
			if targetCallNode == nil {
				break
			}
			_, err = gb.session.CreateDcfgCallEdge(lastDcfgNode, targetCallNode)
			if err != nil {
				return err
			}
			callCount++
		}
		_, err = gb.session.CreateDcfgReturnEdge(lastDcfgNode, callNode)
		if err != nil {
			return err
		}
	}

	// Build all DCFG to AssetFlow edges
	for _, callTrace := range traceResult.CallTraces {
		for _, flow := range callTrace.Flows {
			flowNode := g.nodeMap.GetAssetFlowNode(flow.DcfgId.CallId(), uint64(flow.Index))
			dcfgNode := g.nodeMap.GetDcfgNode(flow.DcfgId)
			if dcfgNode == nil {
				g.logger.Warn("Cannot find DCFG node for flow: ", flow.DcfgId.String())
				continue
			}
			_, err := gb.session.CreateDcfgToAssetFlowEdge(dcfgNode, flowNode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (gb *GraphBuilder) buildDataflowGraph(g *Graph, traceResult *TraceResult) error {
	sourceList := make([]*dataflow.Source, 0)
	sourceSet := mapset.NewSet[*dataflow.Source]()

	// Build dataflow graph for all relevant sources
	// List all sources related to asset flows
	for _, callTrace := range traceResult.CallTraces {
		for _, flow := range callTrace.Flows {
			if flow.AmountTracker == nil {
				continue
			}
			for _, sources := range flow.AmountTracker.Sources {
				for _, source := range sources {
					if !sourceSet.Contains(source) {
						sourceSet.Add(source)
						sourceList = append(sourceList, source)
					}
				}
			}
		}
	}
	// List all sources related to jumpi
	jumpiDataSources := traceResult.GetJumpiDataSource()
	for _, dataSource := range jumpiDataSources {
		for _, sources := range dataSource.Sources {
			for _, source := range sources {
				if !sourceSet.Contains(source) {
					sourceSet.Add(source)
					sourceList = append(sourceList, source)
				}
			}
		}
	}
	// Find all ancestors
	traversalIdx := 0
	for traversalIdx < len(sourceList) {
		curr := sourceList[traversalIdx]
		traversalIdx++
		if curr.Ancestor == nil {
			continue
		}
		for _, sources := range curr.Ancestor.Sources {
			for _, source := range sources {
				if !sourceSet.Contains(source) {
					sourceSet.Add(source)
					sourceList = append(sourceList, source)
				}
			}
		}
	}
	for {
		// Find all storage/balance histories
		for _, source := range sourceList {
			switch source.Loc.(type) {
			case *dataflow.StorageLocation:
				storageLoc := source.Loc.(*dataflow.StorageLocation)
				svs := traceResult.GetStorageHistory(storageLoc)
				for _, sv := range svs {
					for _, dataSource := range sv.GetSources() {
						for _, sources := range dataSource.Sources {
							for _, source := range sources {
								if !sourceSet.Contains(source) {
									sourceList = append(sourceList, source)
									sourceSet.Add(source)
								}
							}
						}
					}
				}
			case *dataflow.BalanceLocation:
				balanceLoc := source.Loc.(*dataflow.BalanceLocation)
				bvs := traceResult.GetBalanceHistory(balanceLoc)
				for _, bv := range bvs {
					for _, dataSource := range bv.GetSources() {
						for _, sources := range dataSource.Sources {
							for _, source := range sources {
								if !sourceSet.Contains(source) {
									sourceList = append(sourceList, source)
									sourceSet.Add(source)
								}
							}
						}
					}
				}
			}
		}
		// If no new source is added, stop
		if traversalIdx == len(sourceList) {
			break
		}
		// Find all ancestors for newly added sources
		for traversalIdx < len(sourceList) {
			curr := sourceList[traversalIdx]
			traversalIdx++
			if curr.Ancestor == nil {
				continue
			}
			for _, sources := range curr.Ancestor.Sources {
				for _, source := range sources {
					if !sourceSet.Contains(source) {
						sourceSet.Add(source)
						sourceList = append(sourceList, source)
					}
				}
			}
		}
	}
	// Create source nodes
	g.logger.Debug("Source count: ", len(sourceList))
	sourceNodeCount := 0
	for _, source := range sourceList {
		if g.nodeMap.GetDataSourceNodeBySource(source) != nil {
			continue
		}
		id := dataflowSourceIdentifier(source)
		sourceNode, err := nodes.CreateDataflowSourceNode(g.g, id, source.Typ.String(), dataflowSourceLocationToString(source.Loc))
		if err != nil {
			return err
		}
		g.nodeMap.RegisterDataSourceNode(source, sourceNode)
		sourceNodeCount++
	}
	g.logger.Debug("Source node count: ", sourceNodeCount)
	// Create edges
	for _, source := range sourceList {
		node := g.nodeMap.GetDataSourceNodeBySource(source)
		// Create dependency edges using ancestor relations
		if source.Ancestor != nil {
			for _, sources := range source.Ancestor.Sources {
				for _, ancestor := range sources {
					_, err := gb.session.CreateDataflowDependencyEdge(g.nodeMap.GetDataSourceNodeBySource(ancestor), node, "val")
					if err != nil {
						return err
					}
				}
			}
		}
		// Create read/write edges for all storage/balance sources
		switch source.Loc.(type) {
		case *dataflow.StorageLocation:
			storageLoc := source.Loc.(*dataflow.StorageLocation)
			history := traceResult.GetStorageHistory(storageLoc)
			sv := history[storageLoc.HistoryId()]
			// Write edge (read edges are handled through asset flow)
			if sv.GetWrite() != nil {
				dcfgNode := g.nodeMap.GetDcfgNode(sv.GetWrite())
				if dcfgNode == nil {
					g.logger.Warn("Cannot find DCFG node for write: ", sv.GetWrite())
				} else {
					_, err := gb.session.CreateDataflowWriteEdge(dcfgNode, node, "val")
					if err != nil {
						return err
					}
				}
			}
			// Transition edges
			if storageLoc.HistoryId() > 0 {
				parent := history[storageLoc.HistoryId()-1]
				for _, dataSource := range parent.GetSources() {
					for _, sources := range dataSource.Sources {
						for _, source := range sources {
							_, err := gb.session.CreateDataflowTransitionEdge(g.nodeMap.GetDataSourceNodeBySource(source), node, "val")
							if err != nil {
								return err
							}
						}
					}
				}
			}
		case *dataflow.BalanceLocation:
			balanceLoc := source.Loc.(*dataflow.BalanceLocation)
			balanceHistory := traceResult.GetBalanceHistory(balanceLoc)
			bv := balanceHistory[balanceLoc.HistoryId()]
			// Read edges
			for _, readCallId := range bv.GetReads() {
				callNode := g.nodeMap.GetContractCallNode(readCallId)
				if callNode == nil {
					g.logger.Warn("Cannot find call node for read: ", readCallId)
					continue
				}
				_, err := gb.session.CreateDataflowReadBalanceEdge(callNode, node, "val")
				if err != nil {
					return err
				}
			}
			// Write edge
			if bv.GetWrite() != nil {
				callNode := g.nodeMap.GetContractCallNode(bv.GetWrite())
				if callNode == nil {
					g.logger.Warn("Cannot find call node for write: ", bv.GetWrite())
				} else {
					_, err := gb.session.CreateDataflowWriteBalanceEdge(callNode, node, "val")
					if err != nil {
						return err
					}
				}
			}
			// Transition edge
			if balanceLoc.HistoryId() > 0 {
				parent := balanceHistory[balanceLoc.HistoryId()-1]
				for _, dataSource := range parent.GetSources() {
					for _, sources := range dataSource.Sources {
						for _, source := range sources {
							_, err := gb.session.CreateDataflowTransitionEdge(g.nodeMap.GetDataSourceNodeBySource(source), node, "val")
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	// Create read edges for all asset flows
	for _, callTrace := range traceResult.CallTraces {
		for _, flow := range callTrace.Flows {
			if flow.AmountTracker == nil {
				continue
			}
			flowNode := g.nodeMap.GetAssetFlowNode(flow.DcfgId.CallId(), uint64(flow.Index))
			for _, sources := range flow.AmountTracker.Sources {
				for _, source := range sources {
					_, err := gb.session.CreateDataflowReadAmountEdge(flowNode, g.nodeMap.GetDataSourceNodeBySource(source), "val")
					if err != nil {
						return err
					}
				}
			}
		}
	}
	// Create control edges for all jumpi
	for dcfgIdStr, dataSource := range jumpiDataSources {
		for _, sources := range dataSource.Sources {
			for _, source := range sources {
				switch source.Loc.(type) {
				// case *dataflow.StorageLocation:
				// 	continue
				case *dataflow.BalanceLocation:
					continue
				default:
				}
				dcfgId, err := types.StringToDcfgId(dcfgIdStr)
				if err != nil {
					return err
				}
				dcfgNode := g.nodeMap.GetDcfgNode(dcfgId)
				if dcfgNode == nil {
					g.logger.Warn("Cannot find DCFG node for jumpi: ", dcfgId)
					continue
				}
				_, err = gb.session.CreateDataflowControlEdge(g.nodeMap.GetDataSourceNodeBySource(source), dcfgNode, "val")
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
