package nodes

import gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"

type Node interface {
	GetVertex() *gremlingo.Vertex
	Label() string
	Id() string
}
