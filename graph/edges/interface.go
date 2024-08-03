package edges

import (
	"github.com/anon/execution-property-graph/graph/nodes"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type Edge interface {
	GetEdge() *gremlingo.Edge
	Label() string
	FromId() string
	ToId() string
	Id() string
}

func EdgeId(label string, from, to nodes.Node) string {
	return label + "#" + from.Id() + "#" + to.Id()
}
