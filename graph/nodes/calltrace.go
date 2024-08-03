package nodes

import (
	"strconv"

	"github.com/anon/execution-property-graph/dynamicEVM/calltrace"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	ContractCallNodeLabel = "contractCall"
	ContractNodeLabel     = "contract"
	AssetFlowNodeLabel    = "assetFlow"
)

type ContractCallNode struct {
	*gremlingo.Vertex

	Address types.Address
	CallId  types.CallId
}

func (n *ContractCallNode) GetVertex() *gremlingo.Vertex {
	return n.Vertex
}

func (n *ContractCallNode) Label() string {
	return ContractCallNodeLabel
}

func (n *ContractCallNode) Id() string {
	return n.Label() + "#" + types.AddressToHex(n.Address) + ":" + n.CallId.String()
}

func getContractCallVertexFromGraph(g *gremlingo.GraphTraversalSource, address types.Address, callId types.CallId) (*gremlingo.Vertex, error) {
	contract, err := g.V().
		Has("address", types.AddressToHex(address)).
		Has("callId", callId.String()).
		HasLabel(ContractCallNodeLabel).
		Next()
	if err != nil {
		return nil, err
	}
	v, err := contract.GetVertex()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func GetContractCallNodeFromGraph(g *gremlingo.GraphTraversalSource, address types.Address, callId types.CallId) (*ContractCallNode, error) {
	v, err := getContractCallVertexFromGraph(g, address, callId)
	if err != nil {
		return nil, err
	}
	return &ContractCallNode{
		Vertex:  v,
		Address: address,
		CallId:  callId,
	}, nil
}

func CreateContractCallNode(g *gremlingo.GraphTraversalSource, address types.Address, callId types.CallId) (*ContractCallNode, error) {
	res, err := g.AddV(ContractCallNodeLabel).
		Property("address", types.AddressToHex(address)).
		Property("callId", callId.String()).
		Next()
	if err != nil {
		return nil, err
	}
	v, err := res.GetVertex()
	if err != nil {
		return nil, err
	}
	return &ContractCallNode{
		Vertex:  v,
		Address: address,
		CallId:  callId,
	}, nil
}

func CreateContractCallNodes(g *gremlingo.GraphTraversalSource, addresses []types.Address, callIds []types.CallId) ([]*ContractCallNode, error) {
	t := g.GetGraphTraversal()
	selects := make([]interface{}, 0, len(addresses))
	for i, addr := range addresses {
		iStr := strconv.Itoa(i)
		selects = append(selects, iStr)
		t.AddV(ContractCallNodeLabel).
			Property("address", types.AddressToHex(addr)).
			Property("callId", callIds[i].String()).
			As(iStr)
	}
	t.Select(selects...)
	res, err := t.Next()
	if err != nil {
		return nil, err
	}
	result := res.GetInterface()
	nodes := make([]*ContractCallNode, 0, len(addresses))
	if vmap, ok := result.(map[interface{}]interface{}); ok {
		for i, addr := range addresses {
			v := vmap[strconv.Itoa(i)].(*gremlingo.Vertex)
			node := &ContractCallNode{
				Vertex:  v,
				Address: addr,
				CallId:  callIds[i],
			}
			nodes = append(nodes, node)
		}
	} else {
		v := result.(*gremlingo.Vertex)
		node := &ContractCallNode{
			Vertex:  v,
			Address: addresses[0],
			CallId:  callIds[0],
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

type AssetFlowNode struct {
	*gremlingo.Vertex

	CallId types.CallId
	Index  uint64
	Flow   *calltrace.Flow
}

func (n *AssetFlowNode) GetVertex() *gremlingo.Vertex {
	return n.Vertex
}

func (n *AssetFlowNode) Label() string {
	return AssetFlowNodeLabel
}

func (n *AssetFlowNode) Id() string {
	return n.Label() + "#" + n.CallId.String() + "#" + strconv.FormatUint(n.Index, 10)
}

func getAssetFlowVertexFromGraph(g *gremlingo.GraphTraversalSource, callId types.CallId, index uint64, flow *calltrace.Flow) (*gremlingo.Vertex, error) {
	contract, err := g.V().
		Property("callId", callId.String()).
		Property("id", strconv.FormatUint(index, 10)).
		Property("asset", types.AddressToHex(flow.Asset)).
		Property("amount", flow.Amount.String()).
		HasLabel(AssetFlowNodeLabel).
		Next()
	if err != nil {
		return nil, err
	}
	v, err := contract.GetVertex()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func GetAssetFlowNodeFromGraph(g *gremlingo.GraphTraversalSource, callId types.CallId, index uint64, flow *calltrace.Flow) (*AssetFlowNode, error) {
	v, err := getAssetFlowVertexFromGraph(g, callId, index, flow)
	if err != nil {
		return nil, err
	}
	return &AssetFlowNode{
		Vertex: v,
		CallId: callId,
		Index:  index,
		Flow:   flow,
	}, nil
}

func CreateAssetFlowNode(g *gremlingo.GraphTraversalSource, callId types.CallId, index uint64, flow *calltrace.Flow) (*AssetFlowNode, error) {
	res, err := g.AddV(AssetFlowNodeLabel).
		Property("callId", callId.String()).
		Property("id", strconv.FormatUint(index, 10)).
		Property("asset", types.AddressToHex(flow.Asset)).
		Property("amount", flow.Amount.String()).
		Next()
	if err != nil {
		return nil, err
	}
	v, err := res.GetVertex()
	if err != nil {
		return nil, err
	}
	return &AssetFlowNode{
		Vertex: v,
		CallId: callId,
		Index:  index,
		Flow:   flow,
	}, nil
}

func CreateAssetFlowNodes(g *gremlingo.GraphTraversalSource, callIds []types.CallId, indices []uint64, flows []*calltrace.Flow) ([]*AssetFlowNode, error) {
	t := g.GetGraphTraversal()
	selects := make([]interface{}, 0, len(callIds))
	for i, callId := range callIds {
		iStr := strconv.Itoa(i)
		selects = append(selects, iStr)
		t.AddV(AssetFlowNodeLabel).
			Property("callId", callId.String()).
			Property("id", strconv.FormatUint(indices[i], 10)).
			Property("asset", types.AddressToHex(flows[i].Asset)).
			Property("amount", flows[i].Amount.String()).
			As(iStr)
	}
	t.Select(selects...)
	res, err := t.Next()
	if err != nil {
		return nil, err
	}
	result := res.GetInterface()
	nodes := make([]*AssetFlowNode, 0, len(callIds))
	if vmap, ok := result.(map[interface{}]interface{}); ok {
		for i, callId := range callIds {
			v := vmap[strconv.Itoa(i)].(*gremlingo.Vertex)
			node := &AssetFlowNode{
				Vertex: v,
				CallId: callId,
				Index:  indices[i],
				Flow:   flows[i],
			}
			nodes = append(nodes, node)
		}
	} else {
		v := result.(*gremlingo.Vertex)
		node := &AssetFlowNode{
			Vertex: v,
			CallId: callIds[0],
			Index:  indices[0],
			Flow:   flows[0],
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
