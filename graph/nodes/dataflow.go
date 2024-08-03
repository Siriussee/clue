package nodes

import (
	"strconv"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const DataSourceNodeLabel = "dataSource"

type DataSourceNode struct {
	*gremlingo.Vertex

	SourceId       string
	SourceType     string
	SourceLocation string
	// Value      string
}

func (n *DataSourceNode) GetVertex() *gremlingo.Vertex {
	return n.Vertex
}

func (n *DataSourceNode) Label() string {
	return DataSourceNodeLabel
}

func (n *DataSourceNode) Id() string {
	return n.Label() + "#" + n.SourceId
}

func getDataflowSourceVertexFromGraph(g *gremlingo.GraphTraversalSource, sourceId, sourceType, sourceLocation string) (*gremlingo.Vertex, error) {
	contract, err := g.V().
		Has("sourceId", sourceId).
		Has("sourceType", sourceType).
		Has("sourceLocation", sourceLocation).
		// Has("value", value).
		HasLabel(DataSourceNodeLabel).
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

func GetDataflowSourceNodeFromGraph(g *gremlingo.GraphTraversalSource, sourceId, sourceType, sourceLocation string) (*DataSourceNode, error) {
	v, err := getDataflowSourceVertexFromGraph(g, sourceId, sourceType, sourceLocation)
	if err != nil {
		return nil, err
	}
	return &DataSourceNode{
		Vertex:         v,
		SourceId:       sourceId,
		SourceType:     sourceType,
		SourceLocation: sourceLocation,
		// Value:      value,
	}, nil
}

func CreateDataflowSourceNode(g *gremlingo.GraphTraversalSource, sourceId, sourceType, sourceLocation string) (*DataSourceNode, error) {
	res, err := g.AddV(DataSourceNodeLabel).
		Property("sourceId", sourceId).
		Property("sourceType", sourceType).
		Property("sourceLocation", sourceLocation).
		// Property("value", value).
		Next()
	if err != nil {
		return nil, err
	}
	v, err := res.GetVertex()
	if err != nil {
		return nil, err
	}
	return &DataSourceNode{
		Vertex:         v,
		SourceId:       sourceId,
		SourceType:     sourceType,
		SourceLocation: sourceLocation,
		// Value:      value,
	}, nil
}

func CreateDataflowSourceNodes(g *gremlingo.GraphTraversalSource, sourceIds, sourceTypes, sourceLocations []string) ([]*DataSourceNode, error) {
	t := g.GetGraphTraversal()
	selects := make([]interface{}, 0, len(sourceTypes))
	for i, typ := range sourceTypes {
		iStr := strconv.Itoa(i)
		selects = append(selects, iStr)
		t.AddV(DataSourceNodeLabel).
			Property("sourceId", sourceIds[i]).
			Property("sourceType", typ).
			Property("sourceLocation", sourceLocations[i]).
			// Property("value", values[i]).
			As(iStr)
	}
	t.Select(selects...)
	res, err := t.Next()
	if err != nil {
		return nil, err
	}
	vmap := res.GetInterface().(map[interface{}]interface{})
	nodes := make([]*DataSourceNode, 0, len(sourceTypes))
	for i, typ := range sourceTypes {
		v := vmap[strconv.Itoa(i)].(*gremlingo.Vertex)
		node := &DataSourceNode{
			Vertex:         v,
			SourceId:       sourceIds[i],
			SourceType:     typ,
			SourceLocation: sourceLocations[i],
			// Value:      values[i],
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
