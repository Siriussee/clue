package edges

import (
	"strconv"

	"github.com/anon/execution-property-graph/dynamicEVM/vm"
	"github.com/anon/execution-property-graph/graph/nodes"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	DcfgJumpEdgeLabel        = "jump"
	DcfgToAssetFlowEdgeLabel = "dcfg_to_asset_flow"
	DcfgReturnLabel          = "dcfg_ret"
	DcfgCallLabel            = "dcfg_call"
)

type DcfgEdge struct {
	*gremlingo.Edge

	From nodes.Node
	// To   *nodes.DcfgNode
	To nodes.Node

	OpCode    vm.OpCode
	Condition bool
}

func (e *DcfgEdge) GetEdge() *gremlingo.Edge {
	return e.Edge
}

func (e *DcfgEdge) Label() string {
	return DcfgJumpEdgeLabel
}

func (e *DcfgEdge) FromId() string {
	return e.From.Id()
}

func (e *DcfgEdge) ToId() string {
	return e.To.Id()
}

func (e *DcfgEdge) Id() string {
	return EdgeId(e.Label(), e.From, e.To)
}

// TODO: change opcode from int to vm.OpCode
func CreateDcfgEdge(g *gremlingo.GraphTraversalSource, from, to *nodes.DcfgNode, opcode int, condition bool) (*DcfgEdge, error) {
	if opcode == -1 {
		opcode = int(vm.JUMP)
	}
	res, err := g.V(from.Vertex).
		AddE(DcfgJumpEdgeLabel).
		To(to.Vertex).
		Property("dcfg:opcode", vm.OpCode(opcode).String()).
		Property("dcfg:condition", strconv.FormatBool(condition)).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DcfgEdge{
		Edge:      e,
		From:      from,
		To:        to,
		OpCode:    vm.OpCode(opcode),
		Condition: condition,
	}, nil
}

func CreateDcfgEdgeTraversal(g *gremlingo.GraphTraversal, from, to *nodes.DcfgNode, opcode int, condition bool) (*gremlingo.GraphTraversal, *DcfgEdge, error) {
	if opcode == -1 {
		opcode = int(vm.JUMP)
	}
	g = g.V(from.Vertex).
		AddE(DcfgJumpEdgeLabel).
		To(to.Vertex).
		Property("dcfg:opcode", vm.OpCode(opcode).String()).
		Property("dcfg:condition", strconv.FormatBool(condition))

	return g, &DcfgEdge{
		From:      from,
		To:        to,
		OpCode:    vm.OpCode(opcode),
		Condition: condition,
	}, nil
}

func CreateDcfgStartEdge(g *gremlingo.GraphTraversalSource, from *nodes.ContractCallNode, to *nodes.DcfgNode) (*DcfgEdge, error) {
	res, err := g.V(from.Vertex).
		AddE(DcfgJumpEdgeLabel).
		To(to.Vertex).
		Property("dcfg:opcode", vm.JUMP.String()).
		Property("dcfg:condition", strconv.FormatBool(true)).
		Next()
	if err != nil {
		return nil, err
	}

	e, err := res.GetEdge()
	if err != nil {
		return nil, err
	}

	return &DcfgEdge{
		Edge:      e,
		From:      from,
		To:        to,
		OpCode:    vm.JUMP,
		Condition: true,
	}, nil
}

func CreateDcfgStartEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.ContractCallNode, to *nodes.DcfgNode) (*gremlingo.GraphTraversal, *DcfgEdge, error) {
	g = g.V(from.Vertex).
		AddE(DcfgJumpEdgeLabel).
		To(to.Vertex).
		Property("dcfg:opcode", vm.JUMP.String()).
		Property("dcfg:condition", strconv.FormatBool(true))

	return g, &DcfgEdge{
		From:      from,
		To:        to,
		OpCode:    vm.JUMP,
		Condition: true,
	}, nil
}

func CreateDcfgReturnEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.DcfgNode, to *nodes.ContractCallNode) (*gremlingo.GraphTraversal, *DcfgEdge, error) {
	g = g.V(from.Vertex).
		AddE(DcfgReturnLabel).
		To(to.Vertex).
		Property("dcfg:opcode", vm.JUMP.String()).
		Property("dcfg:condition", strconv.FormatBool(true))

	return g, &DcfgEdge{
		From:      from,
		To:        to,
		OpCode:    vm.JUMP,
		Condition: true,
	}, nil
}

func CreateDcfgCallEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.DcfgNode, to *nodes.ContractCallNode) (*gremlingo.GraphTraversal, *DcfgEdge, error) {
	g = g.V(from.Vertex).
		AddE(DcfgCallLabel).
		To(to.Vertex).
		Property("dcfg:opcode", vm.JUMP.String()).
		Property("dcfg:condition", strconv.FormatBool(true))

	return g, &DcfgEdge{
		From:      from,
		To:        to,
		OpCode:    vm.JUMP,
		Condition: true,
	}, nil
}

type DcfgToAssetFlowEdge struct {
	*gremlingo.Edge

	From *nodes.DcfgNode
	To   *nodes.AssetFlowNode
}

func (e *DcfgToAssetFlowEdge) GetEdge() *gremlingo.Edge {
	return e.Edge
}

func (e *DcfgToAssetFlowEdge) Label() string {
	return DcfgToAssetFlowEdgeLabel
}

func (e *DcfgToAssetFlowEdge) FromId() string {
	return e.From.Id()
}

func (e *DcfgToAssetFlowEdge) ToId() string {
	return e.To.Id()
}

func (e *DcfgToAssetFlowEdge) Id() string {
	return EdgeId(e.Label(), e.From, e.To)
}

func CreateDcfgToAssetFlowEdgeTraversal(g *gremlingo.GraphTraversal, from *nodes.DcfgNode, to *nodes.AssetFlowNode) (*gremlingo.GraphTraversal, *DcfgToAssetFlowEdge, error) {
	g = g.V(from.Vertex).
		AddE(DcfgToAssetFlowEdgeLabel).
		To(to.Vertex)

	return g, &DcfgToAssetFlowEdge{
		From: from,
		To:   to,
	}, nil
}
