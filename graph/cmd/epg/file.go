package main

import (
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/urfave/cli/v2"
)

const graphPrefix = "/exported-graphs/"

func importGraph(ctx *cli.Context) error {
	remote, err := gremlingo.NewDriverRemoteConnection(ctx.String("remote"))
	if err != nil {
		return err
	}
	defer remote.Close()
	g := gremlingo.Traversal_().WithRemote(remote)

	if err := <-g.V().Drop().Iterate(); err != nil {
		return err
	}

	return <-g.Io(graphPrefix + ctx.String("name") + "." + ctx.String("format")).Read().Iterate()
}

func exportGraph(ctx *cli.Context) error {
	remote, err := gremlingo.NewDriverRemoteConnection(ctx.String("remote"))
	if err != nil {
		return err
	}
	defer remote.Close()
	g := gremlingo.Traversal_().WithRemote(remote)

	if ctx.String("format") != "json" && ctx.String("format") != "xml" {
		return cli.Exit("Invalid format", 1)
	}

	return <-g.Io(graphPrefix + ctx.String("name") + "." + ctx.String("format")).Write().Iterate()
}

func dropGraph(ctx *cli.Context) error {
	remote, err := gremlingo.NewDriverRemoteConnection(ctx.String("remote"))
	if err != nil {
		return err
	}
	defer remote.Close()
	g := gremlingo.Traversal_().WithRemote(remote)

	return <-g.V().Drop().Iterate()
}
