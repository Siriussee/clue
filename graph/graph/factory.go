package graph

import (
	"github.com/anon/execution-property-graph/dynamicEVM/calltrace"
	"github.com/anon/execution-property-graph/graph/edges"
	"github.com/anon/execution-property-graph/graph/nodes"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type GraphFactorySession struct {
	graph         *Graph
	g             *gremlingo.GraphTraversal
	creationBatch int

	creationCount int
}

func NewGraphFactorySession(graph *Graph, creationBatch int) *GraphFactorySession {
	g := graph.g.GetGraphTraversal()
	return &GraphFactorySession{
		graph:         graph,
		g:             g,
		creationBatch: creationBatch,
		creationCount: 0,
	}
}

func (s *GraphFactorySession) addCreation() error {
	if s.creationBatch == 0 {
		return nil
	}
	s.creationCount++
	if s.creationCount >= s.creationBatch {
		s.graph.logger.Debug("submitting graph")
		err := <-s.g.Iterate()
		if err != nil {
			return err
		}
		s.g = s.graph.g.GetGraphTraversal()
		s.creationCount = 0
	}
	return nil
}

func (s *GraphFactorySession) CreateCallEdge(from, to *nodes.ContractCallNode, trace *calltrace.CallTrace) (*edges.CallEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.CallEdgeLabel, from, to); e != nil {
		return e.(*edges.CallEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateCallEdgeTraversal(s.g, from, to, trace)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateAssetFlowEdge(from *nodes.ContractCallNode, to *nodes.AssetFlowNode) (*edges.TransferEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.TransferEdgeLabel, from, to); e != nil {
		return e.(*edges.TransferEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateAssetFlowEdgeTraversal(s.g, from, to)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDcfgEdge(from, to *nodes.DcfgNode, opcode int, condition bool) (*edges.DcfgEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DcfgJumpEdgeLabel, from, to); e != nil {
		return e.(*edges.DcfgEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDcfgEdgeTraversal(s.g, from, to, opcode, condition)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDcfgStartEdge(from *nodes.ContractCallNode, to *nodes.DcfgNode) (*edges.DcfgEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DcfgJumpEdgeLabel, from, to); e != nil {
		return e.(*edges.DcfgEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDcfgStartEdgeTraversal(s.g, from, to)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDcfgReturnEdge(from *nodes.DcfgNode, to *nodes.ContractCallNode) (*edges.DcfgEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DcfgJumpEdgeLabel, from, to); e != nil {
		return e.(*edges.DcfgEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDcfgReturnEdgeTraversal(s.g, from, to)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDcfgCallEdge(from *nodes.DcfgNode, to *nodes.ContractCallNode) (*edges.DcfgEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DcfgJumpEdgeLabel, from, to); e != nil {
		return e.(*edges.DcfgEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDcfgCallEdgeTraversal(s.g, from, to)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDcfgToAssetFlowEdge(from *nodes.DcfgNode, to *nodes.AssetFlowNode) (*edges.DcfgToAssetFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DcfgJumpEdgeLabel, from, to); e != nil {
		return e.(*edges.DcfgToAssetFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDcfgToAssetFlowEdgeTraversal(s.g, from, to)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDataflowControlEdge(from *nodes.DataSourceNode, to *nodes.DcfgNode, value string) (*edges.DataFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DataflowControlEdgeLabel, from, to); e != nil {
		return e.(*edges.DataFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDataflowControlEdgeTraversal(s.g, from, to, value)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDataflowReadAmountEdge(from *nodes.AssetFlowNode, to *nodes.DataSourceNode, value string) (*edges.DataFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DataflowReadEdgeLabel, from, to); e != nil {
		return e.(*edges.DataFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDataflowReadAmountEdgeTraversal(s.g, from, to, value)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDataflowReadBalanceEdge(from *nodes.ContractCallNode, to *nodes.DataSourceNode, value string) (*edges.DataFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DataflowReadEdgeLabel, from, to); e != nil {
		return e.(*edges.DataFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDataflowReadBalanceEdgeTraversal(s.g, from, to, value)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDataflowWriteEdge(from *nodes.DcfgNode, to *nodes.DataSourceNode, value string) (*edges.DataFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DataflowWriteEdgeLabel, from, to); e != nil {
		return e.(*edges.DataFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDataflowWriteEdgeTraversal(s.g, from, to, value)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDataflowWriteBalanceEdge(from *nodes.ContractCallNode, to *nodes.DataSourceNode, value string) (*edges.DataFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DataflowWriteEdgeLabel, from, to); e != nil {
		return e.(*edges.DataFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDataflowWriteBalanceEdgeTraversal(s.g, from, to, value)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDataflowDependencyEdge(from, to *nodes.DataSourceNode, value string) (*edges.DataFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DataflowDependencyEdgeLabel, from, to); e != nil {
		return e.(*edges.DataFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDataflowDependencyEdgeTraversal(s.g, from, to, value)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) CreateDataflowTransitionEdge(from, to *nodes.DataSourceNode, value string) (*edges.DataFlowEdge, error) {
	if e := s.graph.edgeMap.GetEdge(edges.DataflowTransitionEdgeLabel, from, to); e != nil {
		return e.(*edges.DataFlowEdge), nil
	}
	if err := s.addCreation(); err != nil {
		return nil, err
	}
	g, edge, err := edges.CreateDataflowTransitionEdgeTraversal(s.g, from, to, value)
	if err != nil {
		return nil, err
	}
	s.g = g
	s.graph.edgeMap.RegisterEdge(edge)
	return edge, nil
}

func (s *GraphFactorySession) Submit() error {
	s.graph.logger.Debug("submitting graph")
	return <-s.g.Iterate()
}
