package graph

import gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"

type Traversal func(g *gremlingo.GraphTraversalSource) *gremlingo.GraphTraversal

func (g *Graph) Traverse(traversal Traversal) (*gremlingo.Result, error) {
	return traversal(g.g).Next()
}
