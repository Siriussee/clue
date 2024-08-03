package graph

import (
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/sirupsen/logrus"
)

type GraphConfig struct {
	RemoteUrl string
}

type Graph struct {
	config *GraphConfig

	remote *gremlingo.DriverRemoteConnection
	g      *gremlingo.GraphTraversalSource

	logger *logrus.Logger

	nodeMap *NodeMap
	edgeMap *EdgeMap
}

func NewEmptyGraph(config *GraphConfig) (*Graph, error) {
	remote, err := gremlingo.NewDriverRemoteConnection(config.RemoteUrl)
	if err != nil {
		return nil, err
	}
	g := gremlingo.Traversal_().WithRemote(remote)
	return &Graph{
		config:  config,
		remote:  remote,
		g:       g,
		logger:  logrus.New(),
		nodeMap: NewNodeMap(),
		edgeMap: NewEdgeMap(),
	}, nil
}

func (g *Graph) Close() {
	g.remote.Close()
}
