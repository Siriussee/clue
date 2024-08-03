package main

import (
	"os"

	_ "github.com/anon/execution-property-graph/go-ethereum-driver"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var log = logrus.New()

func main() {
	app := &cli.App{
		Name: "epg",
		Commands: []*cli.Command{
			{
				Name: "build",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tx",
						Usage:    "transaction hash to trace",
						Required: true,
					},
					ProfileFlag,
					TimePerformanceFlag,
					GraphDBRemoteFlag,
				},
				Action: buildGraph,
			},
			{
				Name: "trace",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tx",
						Usage:    "transaction hash to trace",
						Required: true,
					},
					ProfileFlag,
					TimePerformanceFlag,
					EthRemoteFlag,
					EthArchiveRemoteFlag,
				},
				Action: traceTx,
			},
			{
				Name: "import",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "graph name to import",
						Value:   "graph",
					},
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "graph format to import (json, xml)",
						Value:   "json",
					},
					GraphDBRemoteFlag,
				},
				Action: importGraph,
			},
			{
				Name: "export",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "graph name to export",
						Value:   "graph",
					},
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "graph format to export (json, xml)",
						Value:   "json",
					},
					GraphDBRemoteFlag,
				},
				Action: exportGraph,
			},
			{
				Name: "drop",
				Flags: []cli.Flag{
					GraphDBRemoteFlag,
				},
				Action: dropGraph,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
