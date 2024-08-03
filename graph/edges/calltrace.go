package edges

import (
	"github.com/anon/execution-property-graph/dynamicEVM/calltrace"
	"github.com/anon/execution-property-graph/dynamicEVM/types"
	"github.com/anon/execution-property-graph/graph/nodes"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	CallEdgeLabel     = "call"
	TransferEdgeLabel = "transfer"
)

type CallEdge struct {
	*gremlingo.Edge

	From *nodes.ContractCallNode
	To   *nodes.ContractCallNode

	CallTrace *calltrace.CallTrace
}

func (e *CallEdge) GetEdge() *gremlingo.Edge {
	return e.Edge
}

func (e *CallEdge) Label() string {
	return CallEdgeLabel
}

func (e *CallEdge) FromId() string {
	return e.From.Id()
}

func (e *CallEdge) ToId() string {
	return e.To.Id()
}

func (e *CallEdge) Id() string {
	return EdgeId(e.Label(), e.From, e.To)
}

func CreateCallEdge(g *gremlingo.GraphTraversalSource, from, to *nodes.ContractCallNode, trace *calltrace.CallTrace) (*CallEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(CallEdgeLabel).
		To(to.Vertex).
		Property("callTrace:id", trace.ID.String()).
		Property("callTrace:from", types.AddressToHex(trace.From)).
		Property("callTrace:to", types.AddressToHex(trace.To)).
		Property("callTrace:type", trace.Type.String()).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &CallEdge{
		Edge:      e,
		From:      from,
		To:        to,
		CallTrace: trace,
	}, nil
}

func CreateCallEdgeTraversal(g *gremlingo.GraphTraversal, from, to *nodes.ContractCallNode, trace *calltrace.CallTrace) (*gremlingo.GraphTraversal, *CallEdge, error) {
	g = g.V(from.Vertex).
		AddE(CallEdgeLabel).
		To(to.Vertex).
		Property("callTrace:id", trace.ID.String()).
		Property("callTrace:from", types.AddressToHex(trace.From)).
		Property("callTrace:to", types.AddressToHex(trace.To)).
		Property("callTrace:type", trace.Type.String())

	return g, &CallEdge{
		From:      from,
		To:        to,
		CallTrace: trace,
	}, nil
}

type TransferEdge struct {
	*gremlingo.Edge

	From *nodes.ContractCallNode
	To   *nodes.AssetFlowNode
}

func (e *TransferEdge) GetEdge() *gremlingo.Edge {
	return e.Edge
}

func (e *TransferEdge) Label() string {
	return TransferEdgeLabel
}

func (e *TransferEdge) FromId() string {
	return e.From.Id()
}

func (e *TransferEdge) ToId() string {
	return e.To.Id()
}

func (e *TransferEdge) Id() string {
	return EdgeId(e.Label(), e.From, e.To)
}

func CreateAssetFlowEdge(g *gremlingo.GraphTraversalSource, from *nodes.ContractCallNode, to *nodes.AssetFlowNode) (*TransferEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(TransferEdgeLabel).
		To(to.Vertex).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &TransferEdge{
		Edge: e,
		From: from,
		To:   to,
	}, nil
}

func CreateAssetFlowEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.ContractCallNode, to *nodes.AssetFlowNode) (*gremlingo.GraphTraversal, *TransferEdge, error) {
	g = g.V(from.Vertex).
		AddE(TransferEdgeLabel).
		To(to.Vertex)

	return g, &TransferEdge{
		From: from,
		To:   to,
	}, nil
}
