package edges

import (
	"github.com/anon/execution-property-graph/graph/nodes"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	DataflowReadEdgeLabel       = "dataflow:read"
	DataflowControlEdgeLabel    = "dataflow:control"
	DataflowWriteEdgeLabel      = "dataflow:write"
	DataflowDependencyEdgeLabel = "dataflow:dependency"
	DataflowTransitionEdgeLabel = "dataflow:transition"
)

type DataFlowEdge struct {
	*gremlingo.Edge

	From nodes.Node
	To   nodes.Node

	EdgeType string
	Value    string
}

func (e *DataFlowEdge) GetEdge() *gremlingo.Edge {
	return e.Edge
}

func (e *DataFlowEdge) Label() string {
	return e.EdgeType
}

func (e *DataFlowEdge) FromId() string {
	return e.From.Id()
}

func (e *DataFlowEdge) ToId() string {
	return e.To.Id()
}

func (e *DataFlowEdge) Id() string {
	return EdgeId(e.Label(), e.From, e.To)
}

func CreateDataflowControlEdge(g *gremlingo.GraphTraversalSource, from *nodes.DataSourceNode, to *nodes.DcfgNode, value string) (*DataFlowEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DataflowControlEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DataFlowEdge{
		Edge:     e,
		From:     from,
		To:       to,
		EdgeType: DataflowControlEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowControlEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.DataSourceNode, to *nodes.DcfgNode, value string) (*gremlingo.GraphTraversal, *DataFlowEdge, error) {
	g.V(from.Vertex).
		AddE(DataflowControlEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value)

	return g, &DataFlowEdge{
		From:     from,
		To:       to,
		EdgeType: DataflowControlEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowReadAmountEdge(g *gremlingo.GraphTraversalSource, from *nodes.AssetFlowNode, to *nodes.DataSourceNode, value string) (*DataFlowEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DataflowReadEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DataFlowEdge{
		Edge:     e,
		From:     from,
		To:       to,
		EdgeType: DataflowReadEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowReadAmountEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.AssetFlowNode, to *nodes.DataSourceNode, value string) (*gremlingo.GraphTraversal, *DataFlowEdge, error) {
	g = g.V(from.Vertex).
		AddE(DataflowReadEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value)

	return g, &DataFlowEdge{
		From:     from,
		To:       to,
		EdgeType: DataflowReadEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowReadBalanceEdge(g *gremlingo.GraphTraversalSource, from *nodes.ContractCallNode, to *nodes.DataSourceNode, value string) (*DataFlowEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DataflowReadEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DataFlowEdge{
		Edge:     e,
		From:     from,
		To:       to,
		EdgeType: DataflowReadEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowReadBalanceEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.ContractCallNode, to *nodes.DataSourceNode, value string) (*gremlingo.GraphTraversal, *DataFlowEdge, error) {
	g = g.V(from.Vertex).
		AddE(DataflowReadEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value)

	return g, &DataFlowEdge{
		From:     from,
		To:       to,
		EdgeType: DataflowReadEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowWriteEdge(g *gremlingo.GraphTraversalSource, from *nodes.DcfgNode, to *nodes.DataSourceNode, value string) (*DataFlowEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DataflowWriteEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DataFlowEdge{
		Edge:     e,
		From:     from,
		To:       to,
		EdgeType: DataflowWriteEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowWriteEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.DcfgNode, to *nodes.DataSourceNode, value string) (*gremlingo.GraphTraversal, *DataFlowEdge, error) {
	g = g.V(from.Vertex).
		AddE(DataflowWriteEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value)

	return g, &DataFlowEdge{
		From:     from,
		To:       to,
		EdgeType: DataflowWriteEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowWriteBalanceEdge(g *gremlingo.GraphTraversalSource, from *nodes.ContractCallNode, to *nodes.DataSourceNode, value string) (*DataFlowEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DataflowWriteEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DataFlowEdge{
		Edge:     e,
		From:     from,
		To:       to,
		EdgeType: DataflowWriteEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowWriteBalanceEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.ContractCallNode, to *nodes.DataSourceNode, value string) (*gremlingo.GraphTraversal, *DataFlowEdge, error) {
	g = g.V(from.Vertex).
		AddE(DataflowWriteEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value)

	return g, &DataFlowEdge{
		From:     from,
		To:       to,
		EdgeType: DataflowWriteEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowDependencyEdge(g *gremlingo.GraphTraversalSource, from, to *nodes.DataSourceNode, value string) (*DataFlowEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DataflowDependencyEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DataFlowEdge{
		Edge:     e,
		From:     from,
		To:       to,
		EdgeType: DataflowDependencyEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowDependencyEdgeTraversal(g *gremlingo.GraphTraversal, from, to *nodes.DataSourceNode, value string) (*gremlingo.GraphTraversal, *DataFlowEdge, error) {
	g = g.V(from.Vertex).
		AddE(DataflowDependencyEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value)

	return g, &DataFlowEdge{
		From:     from,
		To:       to,
		EdgeType: DataflowDependencyEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowTransitionEdge(g *gremlingo.GraphTraversalSource, from, to *nodes.DataSourceNode, value string) (*DataFlowEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DataflowTransitionEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DataFlowEdge{
		Edge:     e,
		From:     from,
		To:       to,
		EdgeType: DataflowTransitionEdgeLabel,
		Value:    value,
	}, nil
}

func CreateDataflowTransitionEdgeTraversal(g *gremlingo.GraphTraversal, from, to *nodes.DataSourceNode, value string) (*gremlingo.GraphTraversal, *DataFlowEdge, error) {
	g = g.V(from.Vertex).
		AddE(DataflowTransitionEdgeLabel).
		To(to.Vertex).
		Property("dataflow:value", value)

	return g, &DataFlowEdge{
		From:     from,
		To:       to,
		EdgeType: DataflowTransitionEdgeLabel,
		Value:    value,
	}, nil
}
