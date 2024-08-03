package graph

import (
	"strconv"

	"github.com/anon/execution-property-graph/dynamicEVM/dataflow"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/graph/edges"
	"github.com/anon/execution-property-graph/graph/nodes"
)

type NodeMap struct {
	rawNodeMap map[interface{}]nodes.Node

	contractCallNodeMap map[string]*nodes.ContractCallNode
	assetFlowNodeMap    map[string]*nodes.AssetFlowNode
	dcfgNodeMap         map[string]*nodes.DcfgNode
	dataSourceNodeMap   map[string]*nodes.DataSourceNode
}

func NewNodeMap() *NodeMap {
	return &NodeMap{
		rawNodeMap:          make(map[interface{}]nodes.Node),
		contractCallNodeMap: make(map[string]*nodes.ContractCallNode),
		assetFlowNodeMap:    make(map[string]*nodes.AssetFlowNode),
		dcfgNodeMap:         make(map[string]*nodes.DcfgNode),
		dataSourceNodeMap:   make(map[string]*nodes.DataSourceNode),
	}
}

func (m *NodeMap) RegisterContractCallNode(node *nodes.ContractCallNode) {
	m.rawNodeMap[node.Vertex.Id] = node
	m.contractCallNodeMap[node.CallId.String()] = node
}

func (m *NodeMap) GetContractCallNode(callId types.CallId) *nodes.ContractCallNode {
	return m.contractCallNodeMap[callId.String()]
}

func (m *NodeMap) RegisterAssetFlowNode(node *nodes.AssetFlowNode) {
	m.rawNodeMap[node.Vertex.Id] = node
	m.assetFlowNodeMap[node.CallId.String()+"/"+strconv.FormatUint(node.Index, 10)] = node
}

func (m *NodeMap) RegisterAssetFlowNodes(nodes []*nodes.AssetFlowNode) {
	for _, node := range nodes {
		m.rawNodeMap[node.Vertex.Id] = node
		m.assetFlowNodeMap[node.CallId.String()+"/"+strconv.FormatUint(node.Index, 10)] = node
	}
}

func (m *NodeMap) GetAssetFlowNode(callId types.CallId, index uint64) *nodes.AssetFlowNode {
	return m.assetFlowNodeMap[callId.String()+"/"+strconv.FormatUint(index, 10)]
}

func (m *NodeMap) RegisterDcfgNode(node *nodes.DcfgNode) {
	m.rawNodeMap[node.Vertex.Id] = node
	m.dcfgNodeMap[node.DcfgId.String()] = node
}

func (m *NodeMap) GetDcfgNode(dcfgId types.DcfgId) *nodes.DcfgNode {
	return m.dcfgNodeMap[dcfgId.String()]
}

func (m *NodeMap) RegisterDataSourceNode(source *dataflow.Source, node *nodes.DataSourceNode) {
	m.rawNodeMap[node.Vertex.Id] = node
	m.dataSourceNodeMap[dataflowSourceIdentifier(source)] = node
}

func (m *NodeMap) GetDataSourceNodeBySource(source *dataflow.Source) *nodes.DataSourceNode {
	return m.dataSourceNodeMap[dataflowSourceIdentifier(source)]
}

func (m *NodeMap) GetDataSourceNode(sourceId string) *nodes.DataSourceNode {
	return m.dataSourceNodeMap[sourceId]
}

type EdgeMap struct {
	edgeMap map[string]edges.Edge
}

func NewEdgeMap() *EdgeMap {
	return &EdgeMap{
		edgeMap: make(map[string]edges.Edge),
	}
}

func (m *EdgeMap) RegisterEdge(edge edges.Edge) {
	m.edgeMap[edge.Id()] = edge
}

func (m *EdgeMap) GetEdge(label string, from, to nodes.Node) edges.Edge {
	return m.edgeMap[edges.EdgeId(label, from, to)]
}

func (m *EdgeMap) HasEdge(label string, from, to nodes.Node) bool {
	_, exists := m.edgeMap[edges.EdgeId(label, from, to)]
	return exists
}
