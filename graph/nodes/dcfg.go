package nodes

import (
	"strconv"

	"github.com/anon/execution-property-graph/dynamicEVM/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const DcfgNodeLabel = "dcfg"

type DcfgNode struct {
	*gremlingo.Vertex

	DcfgId types.DcfgId
	Pc     uint64
}

func (n *DcfgNode) GetVertex() *gremlingo.Vertex {
	return n.Vertex
}

func (n *DcfgNode) Label() string {
	return DcfgNodeLabel
}

func (n *DcfgNode) Id() string {
	return n.Label() + "#" + n.DcfgId.String()
}

func getDcfgVertexFromGraph(g *gremlingo.GraphTraversalSource, dcfgId types.DcfgId, pc uint64) (*gremlingo.Vertex, error) {
	contract, err := g.V().
		Has("dcfgId", dcfgId.String()).
		Has("pc", strconv.FormatUint(pc, 10)).
		HasLabel(DcfgNodeLabel).
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

func GetDcfgNodeFromGraph(g *gremlingo.GraphTraversalSource, dcfgId types.DcfgId, pc uint64) (*DcfgNode, error) {
	v, err := getDcfgVertexFromGraph(g, dcfgId, pc)
	if err != nil {
		return nil, err
	}
	return &DcfgNode{
		Vertex: v,
		DcfgId: dcfgId,
		Pc:     pc,
	}, nil
}

func CreateDcfgNode(g *gremlingo.GraphTraversalSource, dcfgId types.DcfgId, pc uint64) (*DcfgNode, error) {
	res, err := g.AddV(DcfgNodeLabel).
		Property("dcfgId", dcfgId.String()).
		Property("pc", strconv.FormatUint(pc, 10)).
		Next()
	if err != nil {
		return nil, err
	}
	v, err := res.GetVertex()
	if err != nil {
		return nil, err
	}
	return &DcfgNode{
		Vertex: v,
		DcfgId: dcfgId,
		Pc:     pc,
	}, nil
}

func CreateDcfgNodes(g *gremlingo.GraphTraversalSource, dcfgIds []types.DcfgId, pcs []uint64) ([]*DcfgNode, error) {
	if len(dcfgIds) == 0 {
		return nil, nil
	}
	t := g.GetGraphTraversal()
	selects := make([]interface{}, 0, len(dcfgIds))
	for i, id := range dcfgIds {
		iStr := strconv.Itoa(i)
		selects = append(selects, iStr)
		t.AddV(DcfgNodeLabel).
			Property("dcfgId", id.String()).
			Property("pc", strconv.FormatUint(pcs[i], 10)).
			As(iStr)
	}
	t.Select(selects...)
	res, err := t.Next()
	if err != nil {
		return nil, err
	}
	result := res.GetInterface()
	nodes := make([]*DcfgNode, 0, len(dcfgIds))
	if vmap, ok := result.(map[interface{}]interface{}); ok {
		for i, id := range dcfgIds {
			v := vmap[strconv.Itoa(i)].(*gremlingo.Vertex)
			node := &DcfgNode{
				Vertex: v,
				DcfgId: id,
				Pc:     pcs[i],
			}
			nodes = append(nodes, node)
		}
	} else {
		v := result.(*gremlingo.Vertex)
		node := &DcfgNode{
			Vertex: v,
			DcfgId: dcfgIds[0],
			Pc:     pcs[0],
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
